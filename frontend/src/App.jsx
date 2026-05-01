import { useEffect, useMemo, useState } from "react";
import { apiFetch, fetchCurrentUser, getCachedUser, login, logout, signup } from "./auth";

function App() {
  return <Console />;
}

function Console() {
  const [theme, setTheme] = useState(() => localStorage.getItem("coffee.theme") || "dark");
  const [user, setUser] = useState(() => getCachedUser());
  const [products, setProducts] = useState([]);
  const [status, setStatus] = useState("READY");
  const [authMode, setAuthMode] = useState("login");
  const [authOpen, setAuthOpen] = useState(false);
  const [authError, setAuthError] = useState("");
  const [cart, setCart] = useState({});
  const [orderMessage, setOrderMessage] = useState("");

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("coffee.theme", theme);
  }, [theme]);

  useEffect(() => {
    fetchCurrentUser().then(setUser);
  }, []);

  useEffect(() => {
    setStatus(user ? "LOADING MENU" : "LOADING GUEST MENU");
    apiFetch("/products")
      .then(async (response) => {
        if (!response.ok) {
          throw new Error("Menu request failed");
        }
        setProducts(await response.json());
        setStatus(user ? "MENU ONLINE" : "GUEST MENU ONLINE");
      })
      .catch(() => setStatus(user ? "AUTH REQUIRED" : "GUEST RATE LIMITED"));
  }, [user]);

  const hotCount = useMemo(
    () => products.filter((product) => product.category === "hot").length,
    [products]
  );
  const cartItems = useMemo(
    () => products
      .filter((product) => cart[product.id])
      .map((product) => ({ ...product, quantity: cart[product.id] })),
    [cart, products]
  );
  const cartCount = useMemo(
    () => cartItems.reduce((total, item) => total + item.quantity, 0),
    [cartItems]
  );
  const cartTotal = useMemo(
    () => cartItems.reduce((total, item) => total + item.quantity * item.price_in_kurus, 0),
    [cartItems]
  );

  function addToCart(product) {
    setCart((current) => ({
      ...current,
      [product.id]: (current[product.id] || 0) + 1,
    }));
    setOrderMessage("");
    setStatus("ADDED TO CART");
  }

  function setCartQuantity(productId, quantity) {
    setCart((current) => {
      const next = { ...current };
      if (quantity < 1) {
        delete next[productId];
      } else {
        next[productId] = quantity;
      }
      return next;
    });
    setOrderMessage("");
  }

  async function placeOrder() {
    if (cartItems.length === 0) {
      return;
    }

    setStatus("SENDING ORDER");
    setOrderMessage("");
    const response = await apiFetch("/orders", {
      method: "POST",
      body: JSON.stringify({
        items: cartItems.map((item) => ({
          product_id: item.id,
          quantity: item.quantity,
        })),
      }),
    });

    if (!response.ok) {
      setStatus("ORDER FAILED");
      setOrderMessage("Order request failed");
      return;
    }

    const order = await response.json();
    setCart({});
    setStatus("ORDER CREATED");
    setOrderMessage(`ORDER ${order.id.slice(0, 8).toUpperCase()} CREATED`);
  }

  return (
    <main className="console-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">ORDER-SERVICE / SUPERTOKENS</p>
          <h1>Coffee Control</h1>
        </div>
        <div className="toolbar">
          <button type="button" className="icon-button" onClick={() => setTheme(toggleTheme(theme))}>
            {theme === "dark" ? "SUN" : "MOON"}
          </button>
          {user ? (
            <button type="button" onClick={async () => {
              await logout();
              setUser(null);
            }}>LOG OUT</button>
          ) : (
            <button type="button" onClick={() => {
              setAuthMode("login");
              setAuthOpen(true);
            }}>LOG IN</button>
          )}
        </div>
      </header>

      <section className="status-strip" aria-label="System status">
        <span>{status}</span>
        <span>{user ? user.role.toUpperCase() : "GUEST"}</span>
        <span>{cartCount} IN CART</span>
        <span>{hotCount} HOT</span>
      </section>

      {!user && authOpen && (
        <AuthPanel
          mode={authMode}
          error={authError}
          onClose={() => {
            setAuthOpen(false);
            setAuthError("");
          }}
          onModeChange={(mode) => {
            setAuthMode(mode);
            setAuthError("");
          }}
          onSubmit={async (values) => {
            setAuthError("");
            setStatus(authMode === "login" ? "AUTHENTICATING" : "CREATING ACCOUNT");
            try {
              const nextUser = authMode === "login" ? await login(values) : await signup(values);
              setUser(nextUser);
              setAuthOpen(false);
              setStatus("MENU ONLINE");
            } catch (error) {
              setAuthError(error.message);
              setStatus("AUTH FAILED");
            }
          }}
        />
      )}

      <CartPanel
        items={cartItems}
        total={cartTotal}
        message={orderMessage}
        onQuantityChange={setCartQuantity}
        onCheckout={placeOrder}
      />

      <section className="product-grid">
        {products.map((product) => (
          <article key={product.id} className="product-card">
            <div className="product-image">
              <PixelCoffee name={product.name} />
            </div>
            <div className="product-info">
              <p className="chip">{product.category}</p>
              <h2>{product.name}</h2>
              <p>{formatPrice(product.price_in_kurus)}</p>
              <button type="button" onClick={() => addToCart(product)}>ADD</button>
            </div>
          </article>
        ))}
      </section>

      {!user && (
        <section className="guest-panel">
          <span>GUEST ACCESS</span>
          <div className="auth-tabs">
            <button type="button" className={authMode === "login" && authOpen ? "active" : ""} onClick={() => {
              setAuthMode("login");
              setAuthOpen(true);
            }}>LOG IN</button>
            <button type="button" className={authMode === "signup" && authOpen ? "active" : ""} onClick={() => {
              setAuthMode("signup");
              setAuthOpen(true);
            }}>SIGN UP</button>
          </div>
        </section>
      )}
    </main>
  );
}

function CartPanel({ items, total, message, onQuantityChange, onCheckout }) {
  return (
    <section className="cart-panel" aria-label="Cart">
      <div className="cart-head">
        <div>
          <p className="eyebrow">ORDER BUFFER</p>
          <h2>Cart</h2>
        </div>
        <strong>{formatPrice(total)}</strong>
      </div>

      {items.length === 0 ? (
        <p className="cart-empty">NO ITEMS SELECTED</p>
      ) : (
        <div className="cart-items">
          {items.map((item) => (
            <div key={item.id} className="cart-item">
              <span>{item.name}</span>
              <div className="quantity-control">
                <button type="button" aria-label={`Remove one ${item.name}`} onClick={() => onQuantityChange(item.id, item.quantity - 1)}>-</button>
                <strong>{item.quantity}</strong>
                <button type="button" aria-label={`Add one ${item.name}`} onClick={() => onQuantityChange(item.id, item.quantity + 1)}>+</button>
              </div>
              <strong>{formatPrice(item.quantity * item.price_in_kurus)}</strong>
            </div>
          ))}
        </div>
      )}

      <div className="cart-actions">
        {message && <p>{message}</p>}
        <button type="button" disabled={items.length === 0} onClick={onCheckout}>PLACE ORDER</button>
      </div>
    </section>
  );
}

function AuthPanel({ mode, error, onClose, onModeChange, onSubmit }) {
  const [values, setValues] = useState({ email: "", password: "" });
  const isSignup = mode === "signup";

  function update(field, value) {
    setValues((current) => ({ ...current, [field]: value }));
  }

  async function submit(event) {
    event.preventDefault();
    await onSubmit({ email: values.email, password: values.password });
  }

  return (
    <section className="auth-panel">
      <div className="screen-lines" />
      <div className="auth-copy">
        <p className="eyebrow">SECURE TERMINAL</p>
        <h2>{isSignup ? "Create Signal" : "Operator Login"}</h2>
      </div>
      <button type="button" className="auth-close" onClick={onClose}>CLOSE</button>
      <form className="auth-form" onSubmit={submit}>
        <label>
          <span>EMAIL</span>
          <input type="email" value={values.email} onChange={(event) => update("email", event.target.value)} required />
        </label>
        <label>
          <span>PASSWORD</span>
          <input type="password" value={values.password} onChange={(event) => update("password", event.target.value)} minLength={isSignup ? 6 : 1} required />
        </label>
        {error && <p className="auth-error">{error}</p>}
        <button type="submit">{isSignup ? "SIGN UP" : "LOG IN"}</button>
      </form>
      <button type="button" className="text-button" onClick={() => onModeChange(isSignup ? "login" : "signup")}>
        {isSignup ? "HAVE AN ACCOUNT?" : "NEED AN ACCOUNT?"}
      </button>
    </section>
  );
}

function PixelCoffee({ name }) {
  return (
    <div className={`pixel-coffee pixel-${slugify(name)}`} aria-label={`${name} pixel art`}>
      <span className="pixel-steam pixel-steam-a" />
      <span className="pixel-steam pixel-steam-b" />
      <span className="pixel-vessel">
        <span className="pixel-liquid" />
        <span className="pixel-foam" />
        <span className="pixel-syrup" />
        <span className="pixel-ice pixel-ice-a" />
        <span className="pixel-ice pixel-ice-b" />
        <span className="pixel-handle" />
      </span>
      <span className="pixel-saucer" />
    </div>
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

function slugify(value) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "");
}

export default App;
