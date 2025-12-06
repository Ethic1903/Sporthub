const storedGateway = localStorage.getItem("gatewayUrl");
const runtimeGateway = window.__SPORTHUB_GATEWAY__;
const GATEWAY_URL = storedGateway || runtimeGateway || "http://localhost:8080/graphql";
const state = {
  token: localStorage.getItem("sporthubToken") || "",
  currentUser: JSON.parse(localStorage.getItem("sporthubUser") || "null"),
};

class GraphQLClient {
  constructor(endpoint) {
    this.endpoint = endpoint;
  }

  async request(query, variables = {}) {
    const headers = { "Content-Type": "application/json" };
    if (state.token) {
      headers.Authorization = `Bearer ${state.token}`;
    }
    const res = await fetch(this.endpoint, {
      method: "POST",
      headers,
      body: JSON.stringify({ query, variables }),
    });
    const payload = await res.json();
    if (!res.ok || payload.errors) {
      const message = payload?.errors?.[0]?.message || res.statusText;
      throw new Error(message);
    }
    return payload.data;
  }
}

const client = new GraphQLClient(GATEWAY_URL);

const el = (id) => document.getElementById(id);
const tokenPreview = el("token-preview");
const userSpan = el("current-user");
const logBox = el("activity-log");
const facilitiesList = el("facilities-list");
const bookingsList = el("bookings-list");
const connectionStatus = el("connection-status");

function log(message, type = "info") {
  const stamp = new Date().toLocaleTimeString();
  logBox.textContent = `[${stamp}] (${type}) ${message}\n` + logBox.textContent;
}

function setAuthState() {
  tokenPreview.textContent = state.token ? `${state.token.slice(0, 24)}…` : "—";
  userSpan.textContent = state.currentUser ? `${state.currentUser.fullName} (${state.currentUser.role})` : "нет";
}

function saveAuth(token, user) {
  state.token = token || "";
  state.currentUser = user || null;
  if (token) {
    localStorage.setItem("sporthubToken", token);
  } else {
    localStorage.removeItem("sporthubToken");
  }
  if (user) {
    localStorage.setItem("sporthubUser", JSON.stringify(user));
  } else {
    localStorage.removeItem("sporthubUser");
  }
  setAuthState();
}

async function testGateway() {
  try {
    await client.request("query { facilities { id } }", {});
    connectionStatus.textContent = "Gateway reachable";
    connectionStatus.style.background = "rgba(34,197,94,0.2)";
  } catch (err) {
    connectionStatus.textContent = `Gateway error: ${err.message}`;
    connectionStatus.style.background = "rgba(244,63,94,0.2)";
  }
}

function renderFacilities(items) {
  if (!items.length) {
    facilitiesList.innerHTML = `<div class="card">Ничего не найдено</div>`;
    return;
  }
  facilitiesList.innerHTML = items
    .map((f) => {
      const amenities = f.amenities?.length ? f.amenities.join(", ") : "—";
      return `<article class="card">
        <header>
          <strong>${f.name || "Без названия"}</strong>
          <span class="hint">${f.city || "—"}</span>
        </header>
        <p>${f.description || "нет описания"}</p>
        <p class="hint">ID: ${f.id}</p>
        <p class="hint">Тип: ${f.type || "—"}</p>
        <p class="hint">Сервисы: ${amenities}</p>
        <button data-copy="${f.id}">Скопировать ID</button>
      </article>`;
    })
    .join("");
  facilitiesList.querySelectorAll("button[data-copy]").forEach((btn) => {
    btn.addEventListener("click", () => {
      navigator.clipboard.writeText(btn.dataset.copy);
      log(`Скопирован facilityId ${btn.dataset.copy}`);
    });
  });
}

function renderBookings(items) {
  if (!items.length) {
    bookingsList.innerHTML = `<div class="card">Еще нет броней</div>`;
    return;
  }
  bookingsList.innerHTML = items
    .map((b) => {
      return `<article class="card">
        <header class="booking-head">
          <strong>${b.id}</strong>
          <span class="badge">${b.status}</span>
        </header>
        <p>${b.startsAt} → ${b.endsAt}</p>
        <p class="hint">user: ${b.userId || "?"}</p>
        <p class="hint">facility: ${b.facilityId || "?"}</p>
        <button data-confirm="${b.id}">Подтвердить</button>
        <button data-cancel="${b.id}" class="ghost">Отменить</button>
      </article>`;
    })
    .join("");
  bookingsList.querySelectorAll("button[data-confirm]").forEach((btn) =>
    btn.addEventListener("click", () => mutateStatus(btn.dataset.confirm, "confirmBooking"))
  );
  bookingsList.querySelectorAll("button[data-cancel]").forEach((btn) =>
    btn.addEventListener("click", () => mutateStatus(btn.dataset.cancel, "cancelBooking"))
  );
}

async function mutateStatus(id, mutationName) {
  const query = `mutation ($id: ID!) { ${mutationName}(id: $id) { id status } }`;
  try {
    await client.request(query, { id });
    log(`${mutationName} выполнен для ${id}`);
  } catch (err) {
    log(err.message, "error");
  }
}

function buildISO(value) {
  return value ? new Date(value).toISOString() : null;
}

// Event bindings

document.getElementById("register-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = Object.fromEntries(new FormData(e.currentTarget).entries());
  const variables = {
    input: {
      email: data.email,
      password: data.password,
      fullName: data.fullName,
      role: data.role || undefined,
      phone: data.phone || undefined,
    },
  };
  try {
    const res = await client.request(
      `mutation Register($input: RegisterInput!) {
        register(input: $input) { id email fullName }
      }`,
      variables
    );
    log(`Создан пользователь ${res.register.id}`);
    e.currentTarget.reset();
  } catch (err) {
    log(err.message, "error");
  }
});

document.getElementById("login-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = Object.fromEntries(new FormData(e.currentTarget).entries());
  try {
    const res = await client.request(
      `mutation Login($input: LoginInput!) {
        login(input: $input) { token user { id fullName role phone } }
      }`,
      { input: { email: data.email, password: data.password } }
    );
    saveAuth(res.login.token, res.login.user);
    log(`Авторизован ${res.login.user.fullName}`);
  } catch (err) {
    log(err.message, "error");
  }
});

el("logout-btn").addEventListener("click", () => {
  saveAuth("", null);
  log("Токен сброшен");
});

document.getElementById("facility-filter").addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = Object.fromEntries(new FormData(e.currentTarget).entries());
  const variables = {
    city: data.city || undefined,
    type: data.type || undefined,
  };
  try {
    const res = await client.request(
      `query Facilities($city: String, $type: String) {
        facilities(city: $city, type: $type) {
          id name city type description amenities
        }
      }`,
      variables
    );
    renderFacilities(res.facilities || []);
    log(`Загружено площадок: ${res.facilities?.length || 0}`);
  } catch (err) {
    log(err.message, "error");
  }
});

document.getElementById("booking-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!state.token) {
    log("Нужен токен для создания броней", "error");
    return;
  }
  const form = new FormData(e.currentTarget);
  const payload = {
    facilityId: form.get("facilityId"),
    userId: form.get("userId"),
    startsAt: buildISO(form.get("startsAt")),
    endsAt: buildISO(form.get("endsAt")),
    participants: form.get("participants") ? Number(form.get("participants")) : undefined,
    price: form.get("price") ? Number(form.get("price")) : undefined,
    notes: form.get("notes") || undefined,
  };
  try {
    const res = await client.request(
      `mutation CreateBooking($input: BookingInput!) {
        createBooking(input: $input) { id status startsAt endsAt }
      }`,
      { input: payload }
    );
    log(`Создана бронь ${res.createBooking.id}`);
    e.currentTarget.reset();
  } catch (err) {
    log(err.message, "error");
  }
});

document.getElementById("booking-filter").addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = Object.fromEntries(new FormData(e.currentTarget).entries());
  try {
    const res = await client.request(
      `query Bookings($facilityId: ID, $userId: ID) {
        bookings(facilityId: $facilityId, userId: $userId) {
          id status startsAt endsAt facilityId userId
        }
      }`,
      {
        facilityId: data.facilityId || undefined,
        userId: data.userId || undefined,
      }
    );
    renderBookings(res.bookings || []);
    log(`Загружено броней: ${res.bookings?.length || 0}`);
  } catch (err) {
    log(err.message, "error");
  }
});

el("prefill-demo").addEventListener("click", () => {
  const form = document.getElementById("booking-form");
  form.facilityId.value = "fcl-demo";
  form.userId.value = state.currentUser?.id || "usr-demo";
  const start = new Date(Date.now() + 3600 * 1000);
  const end = new Date(start.getTime() + 3600 * 1000);
  form.startsAt.value = start.toISOString().slice(0, 16);
  form.endsAt.value = end.toISOString().slice(0, 16);
  form.participants.value = 4;
  form.price.value = 1000;
  form.notes.value = "training session";
});

el("clear-log").addEventListener("click", () => {
  logBox.textContent = "";
});

// initial load
setAuthState();
testGateway();
document.getElementById("facility-filter").dispatchEvent(new Event("submit"));
document.getElementById("booking-filter").dispatchEvent(new Event("submit"));
