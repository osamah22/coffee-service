const apiUrl = stripTrailingSlash(import.meta.env.VITE_API_URL || "http://localhost:8080");

const storageKeys = {
  session: "coffee.auth.session",
};

const guestSession = {
  token: "",
  user: {
    email: "",
    role: "guest",
  },
};

export async function login(email, password) {
  const credentials = btoa(`${email}:${password}`);
  const response = await fetch(`${apiUrl}/auth/login`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      Authorization: `Basic ${credentials}`,
    },
  });

  if (!response.ok) {
    const message = await readErrorMessage(response, "Login failed");
    throw new Error(message);
  }

  const payload = await response.json();
  const session = {
    token: payload.access_token || "",
    user: {
      email: payload.user?.email || "",
      role: payload.user?.role || "guest",
    },
  };

  persistSession(session);
  return session;
}

export function logout() {
  localStorage.removeItem(storageKeys.session);
}

export function getCachedSession() {
  const raw = localStorage.getItem(storageKeys.session);
  if (!raw) {
    return guestSession;
  }

  try {
    const parsed = JSON.parse(raw);
    return {
      token: parsed.token || "",
      user: {
        ...guestSession.user,
        ...(parsed.user || {}),
      },
    };
  } catch {
    return guestSession;
  }
}

export function getCachedUser() {
  return getCachedSession().user;
}

export async function apiFetch(path, options = {}) {
  const headers = new Headers(options.headers || {});
  const session = getCachedSession();

  headers.set("Accept", "application/json");
  if (session.token) {
    headers.set("Authorization", `Bearer ${session.token}`);
  }
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  return fetch(`${apiUrl}${path}`, { ...options, headers });
}

export function persistSession(session) {
  const nextSession = {
    token: session?.token || "",
    user: {
      ...guestSession.user,
      ...(session?.user || {}),
    },
  };
  localStorage.setItem(storageKeys.session, JSON.stringify(nextSession));
  return nextSession;
}

async function readErrorMessage(response, fallback) {
  try {
    const body = await response.json();
    return body.error || fallback;
  } catch {
    return fallback;
  }
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}
