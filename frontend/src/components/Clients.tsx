import { useState, useEffect } from "react";
import axios from "axios";

export default function Clients() {
  const [clients, setClients] = useState<any[]>([]);
  const [loadingClients, setLoadingClients] = useState(false);
  const [selectedClient, setSelectedClient] = useState<string>("");
  const [bonusData, setBonusData] = useState<any>(null);
  const [spendId, setSpendId] = useState<string>("");
  const [spendAmount, setSpendAmount] = useState<string>("");
  const [spendResult, setSpendResult] = useState<string>("");
  const [bonusLoading, setBonusLoading] = useState(false);
  const [spendLoading, setSpendLoading] = useState(false);
  const [error, setError] = useState<string>("");

  // Load clients list
  const loadClients = async () => {
    setLoadingClients(true);
    setError("");
    try {
      const res = await axios.get("/clients");
      setClients(res.data || []);
    } catch (err: any) {
      setError("Не удалось загрузить список клиентов: " + (err.response?.data || err.message));
    } finally {
      setLoadingClients(false);
    }
  };

  useEffect(() => {
    loadClients();
  }, []);

  // Get client bonus
  const getClientBonus = async () => {
    if (!selectedClient.trim()) return;
    setBonusLoading(true);
    setBonusData(null);
    setError("");
    try {
      const res = await axios.get(`/client_bonus/${selectedClient.trim()}`);
      setBonusData(res.data);
    } catch (err: any) {
      setError(err.response?.data || "Не удалось получить бонусы");
    } finally {
      setBonusLoading(false);
    }
  };

  // Spend bonus
  const handleSpendBonus = async () => {
    if (!spendId.trim() || !spendAmount.trim()) return;
    setSpendLoading(true);
    setSpendResult("");
    setError("");
    try {
      const res = await axios.post(`/spend_bonus/${spendId.trim()}`, {
        amount: parseFloat(spendAmount),
      });
      setSpendResult(res.data.message || "Бонусы успешно списаны");
      setSpendAmount("");
      loadClients();
    } catch (err: any) {
      setError(err.response?.data || "Ошибка при списании бонусов");
    } finally {
      setSpendLoading(false);
    }
  };

  return (
    <div>
      {/* Clients list */}
      <div className="card">
        <div className="section-header">
          <h3>👥 Клиенты</h3>
          <span className="badge">GET /clients</span>
        </div>
        <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 16 }}>
          Список всех клиентов, зарегистрированных в системе.
        </p>

        <button
          className="btn btn-primary"
          onClick={loadClients}
          disabled={loadingClients}
          style={{ marginBottom: 16 }}
        >
          {loadingClients ? "⏳ Загрузка..." : "🔄 Обновить список"}
        </button>

        {clients.length > 0 ? (
          <div style={{ overflowX: "auto" }}>
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Имя</th>
                  <th>Email</th>
                  <th>Телефон</th>
                  <th>Дата создания</th>
                </tr>
              </thead>
              <tbody>
                {clients.map((client: any) => (
                  <tr
                    key={client.id}
                    onClick={() => setSelectedClient(String(client.id))}
                    style={{ cursor: "pointer" }}
                  >
                    <td style={{ fontWeight: 600 }}>{client.id}</td>
                    <td>{client.name}</td>
                    <td>{client.email}</td>
                    <td>{client.phone}</td>
                    <td>{new Date(client.created_at).toLocaleDateString("ru-RU")}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="result-box">
            {loadingClients ? "⏳ Загрузка клиентов..." : "Нет данных о клиентах"}
          </div>
        )}
      </div>

      {/* Bonus info */}
      <div className="card">
        <div className="section-header">
          <h3>💰 Баланс клиента</h3>
          <span className="badge">GET /client_bonus/:id</span>
        </div>
        <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 16 }}>
          Введите ID клиента для просмотра баланса и истории транзакций.
        </p>

        <div className="input-row">
          <div className="input-group">
            <label>ID клиента</label>
            <input
              type="text"
              placeholder="Введите ID клиента"
              value={selectedClient}
              onChange={(e) => setSelectedClient(e.target.value)}
            />
          </div>
          <button
            className="btn btn-primary"
            onClick={getClientBonus}
            disabled={bonusLoading || !selectedClient.trim()}
            style={{ height: "42px" }}
          >
            {bonusLoading ? "⏳ Загрузка..." : "📊 Показать бонусы"}
          </button>
        </div>

        {bonusData && (
          <div>
            <div className="result-box success" style={{ borderLeftColor: "#11998e" }}>
              <strong>Клиент: {bonusData.client?.name}</strong> | Баланс: <strong>{bonusData.balance} ₽</strong>
            </div>

            {bonusData.transactions && bonusData.transactions.length > 0 && (
              <>
                <div className="section-header" style={{ marginTop: 16 }}>
                  <h3>📋 История транзакций</h3>
                  <span className="badge">{bonusData.transactions.length}</span>
                </div>
                <div style={{ overflowX: "auto" }}>
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>ID</th>
                        <th>Тип</th>
                        <th>Сумма (₽)</th>
                        <th>ID заказа</th>
                        <th>Дата</th>
                      </tr>
                    </thead>
                    <tbody>
                      {bonusData.transactions.map((tx: any) => (
                        <tr key={tx.id}>
                          <td>{tx.id}</td>
                          <td>
                            <span
                              style={{
                                padding: "3px 10px",
                                borderRadius: "12px",
                                fontSize: "12px",
                                fontWeight: 600,
                                background:
                                  tx.type === "accrual"
                                    ? "#e6f4ea"
                                    : tx.type === "debit"
                                    ? "#fce8e6"
                                    : "#f0f0f0",
                                color:
                                  tx.type === "accrual"
                                    ? "#1e8e3e"
                                    : tx.type === "debit"
                                    ? "#c5221f"
                                    : "#666",
                              }}
                            >
                              {tx.type === "accrual" ? "Начисление" : tx.type === "debit" ? "Списание" : tx.type}
                            </span>
                          </td>
                          <td style={{ fontWeight: 600 }}>{tx.amount}</td>
                          <td>{tx.order_id}</td>
                          <td>{new Date(tx.created_at).toLocaleString("ru-RU")}</td>
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

      {/* Spend bonus */}
      <div className="card">
        <div className="section-header">
          <h3>✂️ Списать бонусы</h3>
          <span className="badge">POST /spend_bonus/:id</span>
        </div>
        <p style={{ color: "#6c757d", fontSize: 14, marginTop: 0, marginBottom: 16 }}>
          Списать указанную сумму бонусов с баланса клиента.
        </p>

        <div className="input-row">
          <div className="input-group">
            <label>ID клиента</label>
            <input
              type="text"
              placeholder="ID клиента"
              value={spendId}
              onChange={(e) => setSpendId(e.target.value)}
            />
          </div>
          <div className="input-group">
            <label>Сумма списания (₽)</label>
            <input
              type="number"
              placeholder="Например: 500"
              value={spendAmount}
              onChange={(e) => setSpendAmount(e.target.value)}
            />
          </div>
          <button
            className="btn btn-danger"
            onClick={handleSpendBonus}
            disabled={spendLoading || !spendId.trim() || !spendAmount.trim()}
            style={{ height: "42px" }}
          >
            {spendLoading ? "⏳ Списание..." : "✂️ Списать бонусы"}
          </button>
        </div>

        {spendResult && (
          <div className="result-box success">{spendResult}</div>
        )}
      </div>

      {error && (
        <div className="result-box error" style={{ marginTop: 16 }}>
          ❌ {error}
        </div>
      )}
    </div>
  );
}