// App.tsx — головний компонент, health-check бекенду

import { useEffect, useState } from "react";

// Відповідь /ping
interface PingResponse {
  message: string;
  service: string;
}

// Стан запиту
type FetchStatus = "idle" | "loading" | "success" | "error";

// Кореневий компонент
function App() {
  const [status, setStatus] = useState<FetchStatus>("idle");
  const [pingData, setPingData] = useState<PingResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>("");

  // GET /ping
  const fetchPing = async (): Promise<void> => {
    setStatus("loading");
    setPingData(null);
    setErrorMessage("");

    try {
      const response = await fetch("/ping");
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data: PingResponse = await response.json();
      setPingData(data);
      setStatus("success");
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : "Невідома помилка");
      setStatus("error");
    }
  };

  // При монтуванні
  useEffect(() => {
    fetchPing();
  }, []);

  return (
    <div>
      <h1>FinAgent</h1>
      <p>Система обліку фінансів</p>

      <h2>Backend Health Check</h2>

      {status === "loading" && <p>Завантаження...</p>}
      {status === "success" && pingData && (
        <pre>{JSON.stringify(pingData, null, 2)}</pre>
      )}
      {status === "error" && <p>Помилка: {errorMessage}</p>}

      <button onClick={fetchPing} disabled={status === "loading"}>
        Перевірити знову
      </button>
    </div>
  );
}

export default App;
