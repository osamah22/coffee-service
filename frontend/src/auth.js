const issuer = stripTrailingSlash(
  import.meta.env.VITE_AUTHENTIK_ISSUER ||
    "http://localhost:9000/application/o/order-service"
);
const clientId = import.meta.env.VITE_AUTHENTIK_CLIENT_ID || "order-service";
const apiUrl = stripTrailingSlash(import.meta.env.VITE_API_URL || "http://localhost:8080");
const redirectUri = `${window.location.origin}/auth/callback`;
const scope = "openid profile email groups";

const storageKeys = {
  verifier: "coffee.pkce.verifier",
  state: "coffee.pkce.state",
  token: "coffee.auth.token",
  user: "coffee.auth.user",
};

export async function login() {
  const state = crypto.randomUUID();
  const verifier = randomString(64);
  const challenge = await pkceChallenge(verifier);

  sessionStorage.setItem(storageKeys.verifier, verifier);
  sessionStorage.setItem(storageKeys.state, state);

  const params = new URLSearchParams({
    response_type: "code",
    client_id: clientId,
    redirect_uri: redirectUri,
    scope,
    state,
    code_challenge: challenge,
    code_challenge_method: "S256",
  });

  window.location.assign(`${issuer}/authorize/?${params.toString()}`);
}

export async function completeLogin(search) {
  const params = new URLSearchParams(search);
  const code = params.get("code");
  const state = params.get("state");
  const expectedState = sessionStorage.getItem(storageKeys.state);
  const verifier = sessionStorage.getItem(storageKeys.verifier);

  if (!code || !state || state !== expectedState || !verifier) {
    throw new Error("Invalid authentication callback");
  }

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    client_id: clientId,
    code,
    redirect_uri: redirectUri,
    code_verifier: verifier,
  });

  const response = await fetch(`${issuer}/token/`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  });

  if (!response.ok) {
    throw new Error("Token exchange failed");
  }

  const token = await response.json();
  const user = parseUser(token);

  localStorage.setItem(storageKeys.token, JSON.stringify(token));
  localStorage.setItem(storageKeys.user, JSON.stringify(user));
  sessionStorage.removeItem(storageKeys.verifier);
  sessionStorage.removeItem(storageKeys.state);
}

export function logout() {
  localStorage.removeItem(storageKeys.token);
  localStorage.removeItem(storageKeys.user);
  const params = new URLSearchParams({
    client_id: clientId,
    post_logout_redirect_uri: window.location.origin,
  });
  window.location.assign(`${issuer}/end-session/?${params.toString()}`);
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
    localStorage.removeItem(storageKeys.token);
    localStorage.removeItem(storageKeys.user);
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

function parseUser(token) {
  const claims = decodeJWT(token.id_token || token.access_token);
  const groups = normalizeGroups(claims.groups);
  return {
    name: claims.name || claims.preferred_username || claims.email || "Operator",
    email: claims.email || "",
    groups,
    role: groups.some((group) => group.includes("admin")) ? "admin" : "user",
  };
}

function decodeJWT(token) {
  if (!token) {
    return {};
  }
  const [, payload] = token.split(".");
  if (!payload) {
    return {};
  }
  const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
  return JSON.parse(decodeURIComponent(escape(json)));
}

function normalizeGroups(groups) {
  if (Array.isArray(groups)) {
    return groups.map((group) => String(group).toLowerCase());
  }
  if (typeof groups === "string") {
    return groups.split(/[,\s]+/).map((group) => group.toLowerCase());
  }
  return [];
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}

function randomString(length) {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (byte) => (byte % 36).toString(36)).join("");
}

async function pkceChallenge(verifier) {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}
