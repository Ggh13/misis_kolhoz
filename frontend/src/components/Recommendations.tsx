import { useState } from "react";
import axios from "axios";

export default function Recommendations() {
  const [clientId, setClientId] = useState<string>("");
  const [data, setData] = useState<any>(null);
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState(false);

  const fetchRecommendations = async () => {
    if (!clientId.trim()) return;
    setLoading(true);
    setData(null);
    setError("");
    try {
      const res = await axios.get(`/recommendations/${clientId.trim()}`);
      setData(res.data);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") fetchRecommendations();
  };

  return (
    <div className="card">
      <div className="section-header">
        <h3>🤖 Рекомендации товаров</h3>
        <span className="badge">GET /recommendations/:id</span>
      </div>
      <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 20 }}>
        Введите ID клиента, чтобы получить персональные рекомендации товаров на основе его истории покупок.
      </p>

      <div className="input-row">
        <div className="input-group">
          <label>ID клиента</label>
          <input
            type="text"
            placeholder="Например: 1"
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
            onKeyDown={handleKeyDown}
          />
        </div>
        <button
          className="btn btn-primary"
          onClick={fetchRecommendations}
          disabled={loading || !clientId.trim()}
          style={{ height: "42px" }}
        >
          {loading ? "⏳ Загрузка..." : "🔍 Получить рекомендации"}
        </button>
      </div>

      {error && (
        <div className="result-box error">
          ❌ {error}
        </div>
      )}

      {data && (
        <div style={{ marginTop: 20 }}>
          <div className="result-box success" style={{ borderLeftColor: "#11998e" }}>
            <strong>Рекомендации для клиента #{data.client_id}</strong>
          </div>

          {data.recommendations && data.recommendations.length > 0 ? (
            <>
              <p style={{ color: "#6c757d", fontSize: 13, marginTop: 12, marginBottom: 8 }}>
                Найдено рекомендаций: <strong>{data.recommendations.length}</strong>
              </p>
              <div style={{ overflowX: "auto" }}>
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>ID товара</th>
                      <th>Название</th>
                      <th>Категория</th>
                      <th>Ед. измерения</th>
                      <th>Цена (₽)</th>
                      <th>Кол-во</th>
                      <th>ID фермера</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.recommendations.map((rec: any, i: number) => (
                      <tr key={i}>
                        <td style={{ fontWeight: 600 }}>{rec.product_id}</td>
                        <td>{rec.product_name}</td>
                        <td>
                          <span
                            style={{
                              padding: "3px 10px",
                              borderRadius: "12px",
                              fontSize: "12px",
                              fontWeight: 600,
                              background: "#eef2ff",
                              color: "#667eea",
                            }}
                          >
                            {rec.category}
                          </span>
                        </td>
                        <td>{rec.unit}</td>
                        <td>{rec.price}</td>
                        <td>{rec.quantity}</td>
                        <td>{rec.farmer_id}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          ) : (
            <div className="result-box" style={{ borderLeftColor: "#ffc107", marginTop: 8 }}>
              Нет рекомендаций для данного клиента
            </div>
          )}
        </div>
      )}
    </div>
  );
}