import { useEffect, useMemo, useState } from "react";
import { apiFetch, fetchCurrentUser, getCachedUser, login, logout, signup } from "./auth";

function App() {
  return <Console />;
}

function Console() {
  const [theme, setTheme] = useState(() => localStorage.getItem("coffee.theme") || "dark");
  const [user, setUser] = useState(() => getCachedUser());
  const [view, setView] = useState("menu");
  const [products, setProducts] = useState([]);
  const [status, setStatus] = useState("READY");
  const [authMode, setAuthMode] = useState("login");
  const [authOpen, setAuthOpen] = useState(false);
  const [authError, setAuthError] = useState("");
  const [cart, setCart] = useState({});
  const [customerEmail, setCustomerEmail] = useState(() => localStorage.getItem("coffee.customer.email") || "");
  const [orders, setOrders] = useState([]);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [adminOrders, setAdminOrders] = useState([]);
  const [adminLoading, setAdminLoading] = useState(false);
  const [orderMessage, setOrderMessage] = useState("");

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem("coffee.theme", theme);
  }, [theme]);

  useEffect(() => {
    fetchCurrentUser().then((currentUser) => {
      setUser(currentUser);
      if (currentUser?.email) {
        setCustomerEmail(currentUser.email);
        localStorage.setItem("coffee.customer.email", currentUser.email);
      }
    });
  }, []);

  useEffect(() => {
    localStorage.setItem("coffee.customer.email", customerEmail);
  }, [customerEmail]);

  useEffect(() => {
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
  const orderStats = useMemo(() => summarizeOrders(orders), [orders]);
  const adminStats = useMemo(() => summarizeOrders(adminOrders), [adminOrders]);
  const isAdmin = user?.role === "admin";
  const isBarista = user?.role === "barista";
  const canManageOrders = isAdmin || isBarista;
  const canOrder = !isBarista;

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
    const response = await apiFetch("/orders");
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
    const response = await apiFetch(`/orders/${orderId}/${action}`, { method: "POST" });
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
        <div>
          <p className="eyebrow">ORDER-SERVICE / SUPERTOKENS</p>
          <h1>Coffee Control</h1>
        </div>
        <div className="toolbar">
          <button type="button" className={view === "menu" ? "active" : ""} onClick={() => setView("menu")}>
            MENU
          </button>
          {canOrder && (
            <button type="button" className={view === "orders" ? "active" : ""} onClick={() => {
              setView("orders");
              loadOrders();
            }}>
              ORDERS
            </button>
          )}
          {canManageOrders && (
            <button type="button" className={view === "admin" ? "active" : ""} onClick={() => {
              setView("admin");
              loadAdminOrders();
            }}>
              BARISTA
            </button>
          )}
          <button type="button" className="icon-button" onClick={() => setTheme(toggleTheme(theme))}>
            {theme === "dark" ? "SUN" : "MOON"}
          </button>
          {user ? (
            <button type="button" onClick={async () => {
              await logout();
              setUser(null);
              setView("menu");
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
        <span>{canManageOrders ? `${activeOrderCount(adminStats)} ACTIVE` : `${orders.length} PREVIOUS`}</span>
      </section>

      <section className="overview-strip" aria-label="Overview">
        <Metric label="Cart total" value={formatPrice(cartTotal)} />
        <Metric label="Previous orders" value={String(orders.length)} />
        <Metric label="Previous spend" value={formatPrice(orderStats.revenue)} />
        <Metric label={canManageOrders ? "Staff queue" : "Ready drinks"} value={canManageOrders ? String(activeOrderCount(adminStats)) : String(products.length)} />
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
              if (nextUser?.email) {
                setCustomerEmail(nextUser.email);
              }
              setAuthOpen(false);
              setStatus("MENU ONLINE");
            } catch (error) {
              setAuthError(error.message);
              setStatus("AUTH FAILED");
            }
          }}
        />
      )}

      {canOrder && (
        <CartPanel
          items={cartItems}
          total={cartTotal}
          message={orderMessage}
          customerEmail={customerEmail}
          emailLocked={Boolean(user?.email)}
          onEmailChange={setCustomerEmail}
          onQuantityChange={setCartQuantity}
          onCheckout={placeOrder}
        />
      )}

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
          emailLocked={Boolean(user?.email)}
          onEmailChange={setCustomerEmail}
          onRefresh={loadOrders}
        />
      ) : (
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
                {canOrder && <button type="button" onClick={() => addToCart(product)}>ADD</button>}
              </div>
            </article>
          ))}
        </section>
      )}

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

function Metric({ label, value }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
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
