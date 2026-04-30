const apiUrl = stripTrailingSlash(import.meta.env.VITE_API_URL || "http://localhost:8080");

const storageKeys = {
  token: "coffee.auth.token",
  user: "coffee.auth.user",
};

export async function login(credentials) {
  return authenticate("/auth/login", credentials);
}

export async function signup(credentials) {
  return authenticate("/auth/signup", credentials);
}

async function authenticate(path, credentials) {
  const response = await fetch(`${apiUrl}${path}`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const message = await errorMessage(response);
    throw new Error(message);
  }

  const session = await response.json();
  localStorage.setItem(storageKeys.token, JSON.stringify({ access_token: session.access_token }));
  localStorage.setItem(storageKeys.user, JSON.stringify(session.user));
  return session.user;
}

export function logout() {
  localStorage.removeItem(storageKeys.token);
  localStorage.removeItem(storageKeys.user);
}

export function getUser() {
  const raw = localStorage.getItem(storageKeys.user);
  return raw ? JSON.parse(raw) : null;
}

export async function apiFetch(path, options = {}) {
  const token = accessToken();
  const headers = new Headers(options.headers || {});
  headers.set("Accept", "application/json");
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiUrl}${path}`, { ...options, headers });
  if (response.status === 401) {
    logout();
  }
  return response;
}

function accessToken() {
  const raw = localStorage.getItem(storageKeys.token);
  if (!raw) {
    return "";
  }
  const token = JSON.parse(raw);
  return token.access_token || "";
}

async function errorMessage(response) {
  try {
    const body = await response.json();
    return body.error || "Authentication failed";
  } catch {
    return "Authentication failed";
  }
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}
