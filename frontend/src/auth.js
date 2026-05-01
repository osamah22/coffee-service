const apiUrl = stripTrailingSlash(import.meta.env.VITE_API_URL || "http://localhost:8080");

const storageKeys = {
  user: "coffee.auth.user",
};

export async function login(credentials) {
  return authenticate("/auth/signin", credentials);
}

export async function signup(credentials) {
  return authenticate("/auth/signup", credentials);
}

async function authenticate(path, credentials) {
  const response = await fetch(`${apiUrl}${path}`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      formFields: [
        { id: "email", value: credentials.email },
        { id: "password", value: credentials.password },
      ],
    }),
  });

  if (!response.ok) {
    const message = await errorMessage(response);
    throw new Error(message);
  }

  const session = await response.json();
  if (session.status && session.status !== "OK") {
    throw new Error(authStatusMessage(session.status));
  }

  const user = await fetchCurrentUser();
  localStorage.setItem(storageKeys.user, JSON.stringify(user));
  return user;
}

export async function logout() {
  await fetch(`${apiUrl}/auth/signout`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
  }).catch(() => {});
  localStorage.removeItem(storageKeys.user);
}

export function getCachedUser() {
  const raw = localStorage.getItem(storageKeys.user);
  return raw ? JSON.parse(raw) : null;
}

export async function fetchCurrentUser() {
  const response = await sessionFetch("/auth/me");
  if (!response.ok) {
    localStorage.removeItem(storageKeys.user);
    return null;
  }

  const user = await response.json();
  localStorage.setItem(storageKeys.user, JSON.stringify(user));
  return user;
}

export async function apiFetch(path, options = {}) {
  return sessionFetch(path, options);
}

async function sessionFetch(path, options = {}) {
  const headers = new Headers(options.headers || {});
  headers.set("Accept", "application/json");
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  let response = await fetch(`${apiUrl}${path}`, { ...options, credentials: "include", headers });
  if (response.status === 401) {
    const refreshed = await refreshSession();
    if (refreshed) {
      response = await fetch(`${apiUrl}${path}`, { ...options, credentials: "include", headers });
    }
  }
  if (response.status === 401) {
    localStorage.removeItem(storageKeys.user);
  }
  return response;
}

async function refreshSession() {
  const response = await fetch(`${apiUrl}/auth/session/refresh`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
      rid: "session",
    },
  }).catch(() => null);
  return Boolean(response && response.ok);
}

async function errorMessage(response) {
  try {
    const body = await response.json();
    return body.message || body.error || authStatusMessage(body.status) || "Authentication failed";
  } catch {
    return "Authentication failed";
  }
}

function authStatusMessage(status) {
  if (status === "FIELD_ERROR") {
    return "Check your email and password";
  }
  if (status === "WRONG_CREDENTIALS_ERROR") {
    return "Invalid email or password";
  }
  if (status === "EMAIL_ALREADY_EXISTS_ERROR") {
    return "Email already exists";
  }
  return status ? "Authentication failed" : "";
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}
