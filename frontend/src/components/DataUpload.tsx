import { useState } from "react";
import axios from "axios";

export default function DataUpload() {
  const [message, setMessage] = useState<string>("");
  const [loading, setLoading] = useState<string | null>(null);

  const handleUpload = async (endpoint: string) => {
    setLoading(endpoint);
    setMessage("");
    try {
      const res = await axios.post(`/${endpoint}`);
      setMessage(res.data);
    } catch (err: any) {
      setMessage("Ошибка: " + (err.response?.data || err.message));
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="card">
      <div className="section-header">
        <h3>📤 Загрузка данных</h3>
      </div>
      <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 20 }}>
        Нажмите кнопку, чтобы загрузить данные из Excel-файла в базу данных.
      </p>

      <div className="btn-group" style={{ marginBottom: 16 }}>
        <button
          className="btn btn-primary"
          onClick={() => handleUpload("upload_data")}
          disabled={loading !== null}
        >
          {loading === "upload_data"
            ? "⏳ Загрузка данных фермеров..."
            : "📊 Загрузить данные фермеров"}
        </button>

        <button
          className="btn btn-success"
          onClick={() => handleUpload("load_orders")}
          disabled={loading !== null}
        >
          {loading === "load_orders"
            ? "⏳ Загрузка заказов..."
            : "🛒 Загрузить заказы"}
        </button>
      </div>

      {message && (
        <div className={`result-box ${message.startsWith("Ошибка") ? "error" : "success"}`}>
          {message}
        </div>
      )}
    </div>
  );
}