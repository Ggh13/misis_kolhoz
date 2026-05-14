import { useState } from "react";
import axios from "axios";

export default function FarmerSearch() {
  const [farmerId, setFarmerId] = useState<string>("");
  const [data, setData] = useState<any>(null);
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState(false);

  const searchFarmer = async () => {
    if (!farmerId.trim()) return;
    setLoading(true);
    setData(null);
    setError("");
    try {
      const res = await axios.get(`/farmer_data/${farmerId.trim()}`);
      setData(res.data);
    } catch (err: any) {
      setError(err.response?.data || "Фермер не найден");
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") searchFarmer();
  };

  return (
    <div className="card">
      <div className="section-header">
        <h3>🌾 Поиск фермера</h3>
        <span className="badge">GET /farmer_data/:id</span>
      </div>
      <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 20 }}>
        Введите ID фермера (organization_id из Excel), чтобы получить информацию о нём и его продукции.
      </p>

      <div className="input-row">
        <div className="input-group">
          <label>ID фермера</label>
          <input
            type="text"
            placeholder="Например: 12345"
            value={farmerId}
            onChange={(e) => setFarmerId(e.target.value)}
            onKeyDown={handleKeyDown}
          />
        </div>
        <button
          className="btn btn-primary"
          onClick={searchFarmer}
          disabled={loading || !farmerId.trim()}
          style={{ height: "42px" }}
        >
          {loading ? "⏳ Поиск..." : "🔍 Найти"}
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
            <strong>Информация о фермере:</strong>
          </div>

          {data.farmer && (
            <div className="result-box" style={{ marginTop: 8, borderLeftColor: "#667eea" }}>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Поле</th>
                    <th>Значение</th>
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(data.farmer).map(([key, value]) => (
                    <tr key={key}>
                      <td style={{ fontWeight: 600 }}>{key}</td>
                      <td>{String(value)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {data.products && data.products.length > 0 && (
            <>
              <div className="section-header" style={{ marginTop: 16 }}>
                <h3>📦 Продукция</h3>
                <span className="badge">{data.products.length} позиций</span>
              </div>
              <div style={{ overflowX: "auto" }}>
                <table className="data-table">
                  <thead>
                    <tr>
                      {Object.keys(data.products[0]).map((key) => (
                        <th key={key}>{key}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {data.products.map((row: any, i: number) => (
                      <tr key={i}>
                        {Object.values(row).map((val: any, j: number) => (
                          <td key={j}>{String(val)}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}