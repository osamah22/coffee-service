import { useEffect, useMemo, useState } from "react";
import { apiFetch, getCachedSession, getCachedUser, login, logout } from "./auth";

function App() {
  return <Console />;
}

function Console() {
  const [theme, setTheme] = useState(() => localStorage.getItem("coffee.theme") || "dark");
  const [session, setSession] = useState(() => getCachedSession());
  const [user, setUser] = useState(() => getCachedUser());
  const [view, setView] = useState("menu");
  const [products, setProducts] = useState([]);
  const [status, setStatus] = useState(() => (getCachedSession().token ? "READY" : "LOGIN REQUIRED"));
  const [menuFilter, setMenuFilter] = useState("all");
  const [cart, setCart] = useState({});
  const [customerEmail, setCustomerEmail] = useState(() => localStorage.getItem("coffee.customer.email") || "");
  const [orders, setOrders] = useState([]);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [adminOrders, setAdminOrders] = useState([]);
  const [adminLoading, setAdminLoading] = useState(false);
  const [orderMessage, setOrderMessage] = useState("");
  const [authForm, setAuthForm] = useState({ email: user?.email || "customer@example.com", password: "customer123" });
  const [authError, setAuthError] = useState("");

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("coffee.theme", theme);
  }, [theme]);

  useEffect(() => {
    localStorage.setItem("coffee.customer.email", customerEmail);
  }, [customerEmail]);

  useEffect(() => {
    if (!session.token) {
      setProducts([]);
      setStatus("LOGIN REQUIRED");
      return;
    }

    apiFetch("/products")
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(response.status === 401 ? "LOGIN REQUIRED" : "Menu request failed");
        }
        setProducts(await response.json());
        setStatus(user?.role === "user" || user?.role === "admin" ? "MENU ONLINE" : "STAFF MENU ONLINE");
      })
      .catch((error) => setStatus(error.message === "LOGIN REQUIRED" ? "LOGIN REQUIRED" : "MENU REQUEST FAILED"));
  }, [session.token, user?.role]);

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
  const filteredProducts = useMemo(
    () => products.filter((product) => menuFilter === "all" || product.category === menuFilter),
    [menuFilter, products]
  );
  const orderStats = useMemo(() => summarizeOrders(orders), [orders]);
  const adminStats = useMemo(() => summarizeOrders(adminOrders), [adminOrders]);
  const isAuthenticated = Boolean(session.token);
  const isAdmin = user?.role === "admin";
  const isStaff = user?.role === "barista";
  const canManageOrders = isAdmin || isStaff;
  const canOrder = isAuthenticated && !isStaff;

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

  function updateCustomerEmail(email) {
    setCustomerEmail(email);
  }

  async function handleLogin(nextEmail = authForm.email, nextPassword = authForm.password) {
    try {
      setStatus("AUTHENTICATING");
      setAuthError("");
      const nextSession = await login(nextEmail.trim(), nextPassword);
      setSession(nextSession);
      setUser(nextSession.user);
      setCustomerEmail(nextSession.user.email || customerEmail);
      setAuthForm({ email: nextSession.user.email || nextEmail, password: nextPassword });
      setView(nextSession.user.role === "barista" ? "admin" : "menu");
      setStatus(`${nextSession.user.role.toUpperCase()} LOGIN OK`);
    } catch (error) {
      setAuthError(error.message || "Login failed");
      setStatus("LOGIN FAILED");
    }
  }

  function handleLogout() {
    logout();
    const nextSession = getCachedSession();
    setSession(nextSession);
    setUser(nextSession.user);
    setProducts([]);
    setOrders([]);
    setAdminOrders([]);
    setCart({});
    setView("menu");
    setAuthError("");
    setStatus("LOGIN REQUIRED");
  }

  async function placeOrder() {
    if (cartItems.length === 0) {
      return;
    }
    if (!customerEmail.trim()) {
      setStatus("EMAIL REQUIRED");
      setOrderMessage("Enter an email for the receipt");
      return;
    }

    setStatus("SENDING ORDER");
    setOrderMessage("");
    const response = await apiFetch("/orders", {
      method: "POST",
      body: JSON.stringify({
        customer_email: customerEmail.trim(),
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
    setOrders((current) => [order, ...current.filter((item) => item.id !== order.id)]);
    setView("orders");
    setStatus("ORDER CREATED");
    setOrderMessage(`ORDER ${order.id.slice(0, 8).toUpperCase()} CREATED`);
  }

  async function loadOrders() {
    if (!customerEmail.trim()) {
      setStatus("EMAIL REQUIRED");
      setOrderMessage("Enter an email to find orders");
      return;
    }

    setOrdersLoading(true);
    setStatus("LOADING ORDERS");
    const response = await apiFetch(`/orders/mine?email=${encodeURIComponent(customerEmail.trim())}`);
    setOrdersLoading(false);

    if (!response.ok) {
      setStatus("ORDERS FAILED");
      return;
    }

    setOrders(await response.json());
    setStatus("ORDERS ONLINE");
  }

  async function loadAdminOrders() {
    if (!canManageOrders) {
      setStatus("STAFF REQUIRED");
      return;
    }

    setAdminLoading(true);
    setStatus("LOADING QUEUE");
    const response = await apiFetch("/staff/orders");
    setAdminLoading(false);

    if (!response.ok) {
      setStatus(response.status === 403 ? "STAFF DENIED" : "QUEUE FAILED");
      return;
    }

    setAdminOrders(await response.json());
    setStatus("QUEUE ONLINE");
  }

  async function updateAdminOrder(orderId, action) {
    setAdminLoading(true);
    setStatus(`${action.toUpperCase()} ORDER`);
    const response = await apiFetch(`/staff/orders/${orderId}/${action}`, { method: "POST" });
    setAdminLoading(false);

    if (!response.ok) {
      setStatus("ORDER UPDATE FAILED");
      return;
    }

    const updatedOrder = await response.json();
    setAdminOrders((current) => replaceOrder(current, updatedOrder));
    setOrders((current) => replaceOrder(current, updatedOrder));
    setStatus("ORDER UPDATED");
  }

  return (
    <main className="console-shell">
      <header className="topbar">
        <div className="brand-block">
          <p className="eyebrow">ORDER-SERVICE / BASIC AUTH + JWT</p>
          <h1>Coffee Control</h1>
          <span>{isAuthenticated ? `${user.email} / ${user.role}` : "Use one of the demo accounts to start a session"}</span>
        </div>
        <div className="topbar-actions" aria-label="Primary actions">
          <div className="task-tabs" aria-label="Customer tasks">
            <button type="button" className={view === "menu" ? "active" : ""} onClick={() => setView("menu")}>
              Order
            </button>
            {canOrder && !isStaff && (
              <button type="button" className={view === "orders" ? "active" : ""} onClick={() => {
                setView("orders");
                loadOrders();
              }}>
                My Orders
              </button>
            )}
          </div>
          {canManageOrders && (
            <div className="task-tabs staff-tabs" aria-label="Staff tasks">
              <button type="button" className={view === "admin" ? "active" : ""} onClick={() => {
                setView("admin");
                loadAdminOrders();
              }}>
                Staff Queue
              </button>
            </div>
          )}
          <div className="utility-actions">
            <button type="button" className="icon-button" onClick={() => setTheme(toggleTheme(theme))}>
              {theme === "dark" ? "SUN" : "MOON"}
            </button>
            <button type="button" onClick={handleLogout}>Logout</button>
          </div>
        </div>
      </header>

      <section className="status-strip" aria-label="System status">
        <span>{status}</span>
        <span>{user.role.toUpperCase()}</span>
        <span>{cartCount} IN CART</span>
        <span>{canManageOrders ? `${activeOrderCount(adminStats)} ACTIVE` : `${orders.length} PREVIOUS`}</span>
      </section>

      <TaskSummary
        view={view}
        cartTotal={cartTotal}
        cartCount={cartCount}
        productCount={products.length}
        orderStats={orderStats}
        adminStats={adminStats}
        canManageOrders={canManageOrders}
      />

      {canManageOrders && view === "admin" ? (
        <OperationsDashboard
          orders={adminOrders}
          loading={adminLoading}
          stats={adminStats}
          role={user.role}
          onRefresh={loadAdminOrders}
          onUpdateStatus={updateAdminOrder}
        />
      ) : view === "orders" ? (
          <OrdersPanel
            orders={orders}
            loading={ordersLoading}
            stats={orderStats}
            customerEmail={customerEmail}
            emailLocked={user.role === "customer"}
            onEmailChange={updateCustomerEmail}
            onRefresh={loadOrders}
          />
      ) : (
        <section className="order-workspace" aria-label="Order workspace">
          <div className="menu-panel">
            <div className="menu-head">
              <div>
                <p className="eyebrow">MENU</p>
                <h2>Pick Drinks</h2>
              </div>
              <div className="category-tabs" aria-label="Menu filter">
                {["all", "hot", "cold"].map((category) => (
                  <button
                    key={category}
                    type="button"
                    className={menuFilter === category ? "active" : ""}
                    onClick={() => setMenuFilter(category)}
                  >
                    {category}
                  </button>
                ))}
              </div>
            </div>
            <div className="product-grid">
              {filteredProducts.map((product) => (
                <article key={product.id} className="product-card">
                  <div className="product-image">
                    <PixelCoffee name={product.name} />
                  </div>
                  <div className="product-info">
                    <p className="chip">{product.category}</p>
                    <h2>{product.name}</h2>
                    <p className="product-price">{formatPrice(product.price_in_kurus)}</p>
                    {canOrder && <button type="button" onClick={() => addToCart(product)}>Add</button>}
                  </div>
                </article>
              ))}
            </div>
          </div>
          {canOrder && (
            <CartPanel
              items={cartItems}
              total={cartTotal}
              message={orderMessage}
              customerEmail={customerEmail}
              emailLocked={user.role === "user"}
              onEmailChange={updateCustomerEmail}
              onQuantityChange={setCartQuantity}
              onCheckout={placeOrder}
            />
          )}
        </section>
      )}

      <section className="guest-panel">
        <AuthPanel
          authError={authError}
          isAuthenticated={isAuthenticated}
          authForm={authForm}
          onChange={setAuthForm}
          onLogin={handleLogin}
          onLogout={handleLogout}
          user={user}
        />
      </section>
    </main>
  );
}

function TaskSummary({ view, cartTotal, cartCount, productCount, orderStats, adminStats, canManageOrders }) {
  if (view === "admin" && canManageOrders) {
    return (
      <section className="overview-strip" aria-label="Staff queue summary">
        <Metric label="Preparing" value={String(adminStats.preparing)} />
        <Metric label="Ready" value={String(adminStats.ready)} />
        <Metric label="Completed" value={String(adminStats.completed)} />
        <Metric label="Queue revenue" value={formatPrice(adminStats.revenue)} />
      </section>
    );
  }

  if (view === "orders") {
    return (
      <section className="overview-strip" aria-label="Order history summary">
        <Metric label="Orders" value={String(orderStats.count)} />
        <Metric label="Preparing" value={String(orderStats.preparing)} />
        <Metric label="Ready" value={String(orderStats.ready)} />
        <Metric label="Total spend" value={formatPrice(orderStats.revenue)} />
      </section>
    );
  }

  return (
    <section className="overview-strip" aria-label="Ordering summary">
      <Metric label="Drinks" value={String(productCount)} />
      <Metric label="Cart items" value={String(cartCount)} />
      <Metric label="Cart total" value={formatPrice(cartTotal)} />
      <Metric label="Minimum coffee" value={formatPrice(12000)} />
    </section>
  );
}

function Metric({ label, value }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function AuthPanel({ authError, isAuthenticated, authForm, onChange, onLogin, onLogout, user }) {
  const demoAccounts = [
    { label: "User", email: "customer@example.com", password: "customer123" },
    { label: "Barista", email: "barista@coffee.local", password: "barista123" },
    { label: "Admin", email: "admin@coffee.local", password: "admin123" },
  ];

  return (
    <>
      <span>DEMO CREDENTIALS / JWT SESSION</span>
      <div className="auth-tabs">
        <span>{isAuthenticated ? `Bearer token active for ${user.email}` : "No active bearer token"}</span>
        <span>Role: {user.role}</span>
      </div>
      <div className="task-tabs" aria-label="Demo credentials">
        {demoAccounts.map((account) => (
          <button
            key={account.email}
            type="button"
            className={authForm.email === account.email ? "active" : ""}
            onClick={() => {
              onChange({ email: account.email, password: account.password });
              onLogin(account.email, account.password);
            }}
          >
            {account.label}
          </button>
        ))}
      </div>
      <label className="email-field">
        <span>LOGIN EMAIL</span>
        <input
          type="email"
          value={authForm.email}
          onChange={(event) => onChange((current) => ({ ...current, email: event.target.value }))}
          placeholder="customer@example.com"
        />
      </label>
      <label className="email-field">
        <span>LOGIN PASSWORD</span>
        <input
          type="password"
          value={authForm.password}
          onChange={(event) => onChange((current) => ({ ...current, password: event.target.value }))}
          placeholder="customer123"
        />
      </label>
      <div className="cart-actions">
        {authError && <p>{authError}</p>}
        <button type="button" onClick={() => onLogin()}>
          {isAuthenticated ? "REFRESH SESSION" : "LOGIN"}
        </button>
        {isAuthenticated && <button type="button" onClick={onLogout}>CLEAR SESSION</button>}
      </div>
    </>
  );
}

function CartPanel({
  items,
  total,
  message,
  customerEmail,
  emailLocked,
  onEmailChange,
  onQuantityChange,
  onCheckout,
}) {
  return (
    <section className="cart-panel" aria-label="Cart">
      <div className="cart-head">
        <div>
          <p className="eyebrow">ORDER BUFFER</p>
          <h2>Cart</h2>
        </div>
        <strong>{formatPrice(total)}</strong>
      </div>

      <label className="email-field">
        <span>RECEIPT EMAIL</span>
        <input
          type="email"
          value={customerEmail}
          disabled={emailLocked}
          onChange={(event) => onEmailChange(event.target.value)}
          placeholder="operator@example.com"
        />
      </label>

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

function OrdersPanel({ orders, loading, stats, customerEmail, emailLocked, onEmailChange, onRefresh }) {
  return (
    <section className="orders-panel" aria-label="Orders">
      <div className="orders-head">
        <div>
          <p className="eyebrow">PREVIOUS ORDERS</p>
          <h2>Order History</h2>
        </div>
        <button type="button" onClick={onRefresh}>{loading ? "SCANNING" : "REFRESH"}</button>
      </div>

      <div className="summary-grid">
        <Metric label="Total orders" value={String(stats.count)} />
        <Metric label="Preparing" value={String(stats.preparing)} />
        <Metric label="Ready" value={String(stats.ready)} />
        <Metric label="Total spend" value={formatPrice(stats.revenue)} />
      </div>

      <label className="email-field">
        <span>ORDER EMAIL</span>
        <input
          type="email"
          value={customerEmail}
          disabled={emailLocked}
          onChange={(event) => onEmailChange(event.target.value)}
          placeholder="operator@example.com"
        />
      </label>

      {orders.length === 0 ? (
        <p className="cart-empty">{loading ? "LOADING ORDERS" : "NO ORDERS FOUND"}</p>
      ) : (
        <div className="order-list">
          {orders.map((order) => (
            <article key={order.id} className="order-card">
              <div className="order-card-head">
                <strong>#{order.id.slice(0, 8).toUpperCase()}</strong>
                <span>{order.status.toUpperCase()}</span>
                <strong>{formatPrice(order.total)}</strong>
              </div>
              <div className="order-lines">
                {order.items.map((item) => (
                  <div key={`${order.id}-${item.product_id}`} className="order-line">
                    <span>{item.quantity}x {item.product_name || item.product_id.slice(0, 8).toUpperCase()}</span>
                    <strong>{formatPrice(item.quantity * item.price_in_kurus)}</strong>
                  </div>
                ))}
              </div>
              <time dateTime={order.created_at}>{formatDate(order.created_at)}</time>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function OperationsDashboard({ orders, loading, stats, role, onRefresh, onUpdateStatus }) {
  const sortedOrders = useMemo(() => sortOperationalOrders(orders), [orders]);

  return (
    <section className="admin-panel" aria-label="Barista dashboard">
      <div className="admin-hero">
        <div>
          <p className="eyebrow">{role === "admin" ? "ADMIN DASHBOARD" : "BARISTA DASHBOARD"}</p>
          <h2>Barista Queue</h2>
        </div>
        <button type="button" onClick={onRefresh}>{loading ? "SYNCING" : "SYNC"}</button>
      </div>

      <div className="summary-grid">
        <Metric label="Preparing" value={String(stats.preparing)} />
        <Metric label="Ready" value={String(stats.ready)} />
        <Metric label="Completed" value={String(stats.completed)} />
        <Metric label="Revenue" value={formatPrice(stats.revenue)} />
      </div>

      {orders.length === 0 ? (
        <p className="cart-empty">{loading ? "LOADING QUEUE" : "NO ORDERS IN SYSTEM"}</p>
      ) : (
        <div className="admin-order-list">
          {sortedOrders.map((order) => {
            const isFinal = order.status === "completed" || order.status === "cancelled";
            const canMarkReady = order.status === "preparing";
            const canComplete = order.status === "ready";

            return (
              <article key={order.id} className={`admin-order-card status-${order.status}`}>
                <div className="admin-order-main">
                  <div>
                    <strong>#{order.id.slice(0, 8).toUpperCase()}</strong>
                    <span>{order.customer_email}</span>
                  </div>
                  <div>
                    <span>{formatDate(order.created_at)}</span>
                    <strong>{formatPrice(order.total)}</strong>
                  </div>
                  <p>{order.items.map((item) => `${item.quantity}x ${item.product_name || item.product_id.slice(0, 8).toUpperCase()}`).join(" / ")}</p>
                </div>
                <div className="admin-order-actions">
                  <span>{order.status.toUpperCase()}</span>
                  <button type="button" disabled={!canMarkReady || loading} onClick={() => onUpdateStatus(order.id, "ready")}>
                    READY
                  </button>
                  <button type="button" disabled={!canComplete || loading} onClick={() => onUpdateStatus(order.id, "complete")}>
                    COMPLETE
                  </button>
                  <button type="button" disabled={isFinal || loading} onClick={() => onUpdateStatus(order.id, "cancel")}>
                    CANCEL
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
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

function formatDate(value) {
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function summarizeOrders(orderList) {
  return orderList.reduce((stats, order) => {
    const status = order.status || "preparing";
    return {
      count: stats.count + 1,
      preparing: stats.preparing + (status === "preparing" ? 1 : 0),
      completed: stats.completed + (status === "completed" ? 1 : 0),
      cancelled: stats.cancelled + (status === "cancelled" ? 1 : 0),
      ready: stats.ready + (status === "ready" ? 1 : 0),
      revenue: stats.revenue + (status === "cancelled" ? 0 : order.total),
    };
  }, {
    count: 0,
    preparing: 0,
    ready: 0,
    completed: 0,
    cancelled: 0,
    revenue: 0,
  });
}

function replaceOrder(orderList, order) {
  return orderList.map((item) => (item.id === order.id ? order : item));
}

function activeOrderCount(stats) {
  return stats.preparing + stats.ready;
}

function sortOperationalOrders(orderList) {
  const priority = {
    ready: 0,
    preparing: 1,
    completed: 2,
    cancelled: 3,
  };
  return [...orderList].sort((left, right) => {
    const statusDiff = (priority[left.status] ?? 9) - (priority[right.status] ?? 9);
    if (statusDiff !== 0) {
      return statusDiff;
    }
    return new Date(right.created_at).getTime() - new Date(left.created_at).getTime();
  });
}

function slugify(value) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "");
}

export default App;
