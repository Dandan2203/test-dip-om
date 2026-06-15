import axios from "axios";

const api = axios.create({ baseURL: "/api" });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// 401 на захищеному запиті — токен протух, скидаємо сесію та йдемо на /login.
// Виняток: 401 від самих /auth/* (невірний пароль) — НЕ редіректимо, щоб форма
// лишилась і показала помилку.
api.interceptors.response.use(
  (r) => r,
  (err) => {
    const url: string = err.config?.url ?? "";
    const isAuthRequest = url.includes("/auth/");
    if (err.response?.status === 401 && !isAuthRequest) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(err);
  }
);

export default api;
