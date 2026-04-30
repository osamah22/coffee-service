import { useEffect, useMemo, useState } from "react";
import { Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { apiFetch, completeLogin, getUser, login, logout } from "./auth";

function App() {
  return (
    <Routes>
      <Route path="/auth/callback" element={<AuthCallback />} />
      <Route path="*" element={<Console />} />
    </Routes>
  );
}

function AuthCallback() {
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    completeLogin(location.search)
      .then(() => navigate("/", { replace: true }))
      .catch(() => navigate("/", { replace: true, state: { authError: true } }));
  }, [location.search, navigate]);

  return <main className="terminal-screen">SYNCING SESSION...</main>;
}

function Console() {
  const [theme, setTheme] = useState(() => localStorage.getItem("coffee.theme") || "dark");
  const [user, setUser] = useState(() => getUser());
  const [products, setProducts] = useState([]);
  const [status, setStatus] = useState("READY");

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("coffee.theme", theme);
  }, [theme]);

  useEffect(() => {
    if (!user) {
      return;
    }
    setStatus("LOADING MENU");
    apiFetch("/products")
      .then(async (response) => {
        if (!response.ok) {
          throw new Error("Menu request failed");
        }
        setProducts(await response.json());
        setStatus("MENU ONLINE");
      })
      .catch(() => setStatus("AUTH REQUIRED"));
  }, [user]);

  const hotCount = useMemo(
    () => products.filter((product) => product.category === "hot").length,
    [products]
  );

  return (
    <main className="console-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">ORDER-SERVICE / AUTHENTIK</p>
          <h1>Coffee Control</h1>
        </div>
        <div className="toolbar">
          <button type="button" className="icon-button" onClick={() => setTheme(toggleTheme(theme))}>
            {theme === "dark" ? "SUN" : "MOON"}
          </button>
          {user ? (
            <button type="button" onClick={logout}>LOG OUT</button>
          ) : (
            <button type="button" onClick={login}>LOG IN</button>
          )}
        </div>
      </header>

      <section className="status-strip" aria-label="System status">
        <span>{status}</span>
        <span>{user ? user.role.toUpperCase() : "GUEST"}</span>
        <span>{products.length} ITEMS</span>
        <span>{hotCount} HOT</span>
      </section>

      {!user ? (
        <section className="login-panel">
          <div className="screen-lines" />
          <h2>ACCESS LOCKED</h2>
          <p>Authenticate through Authentik to load the menu and role gates.</p>
          <button type="button" onClick={login}>START AUTH</button>
        </section>
      ) : (
        <section className="product-grid">
          {products.map((product) => (
            <article key={product.id} className="product-card">
              <div className="product-image">
                <span>{product.image_path}</span>
              </div>
              <div className="product-info">
                <p className="chip">{product.category}</p>
                <h2>{product.name}</h2>
                <p>{formatPrice(product.price_in_kurus)}</p>
              </div>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}

function toggleTheme(theme) {
  return theme === "dark" ? "light" : "dark";
}

function formatPrice(kurus) {
  return new Intl.NumberFormat("tr-TR", {
    style: "currency",
    currency: "TRY",
  }).format(kurus / 100);
}

export default App;
