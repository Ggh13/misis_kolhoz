import { useState, useEffect } from "react";
import axios from "axios";
import "./App.css";
import DataUpload from "./components/DataUpload";
import FarmerSearch from "./components/FarmerSearch";
import Clients from "./components/Clients";
import Recommendations from "./components/Recommendations";

type ActiveTab = "upload" | "farmers" | "clients" | "recommendations";

function App() {
  const [activeTab, setActiveTab] = useState<ActiveTab>("upload");
  const [healthStatus, setHealthStatus] = useState<boolean | null>(null);

  useEffect(() => {
    axios
      .get("/health")
      .then(() => setHealthStatus(true))
      .catch(() => setHealthStatus(false));
  }, []);

  const tabs: { id: ActiveTab; label: string; icon: string }[] = [
    { id: "upload", label: "Загрузка данных", icon: "📤" },
    { id: "farmers", label: "Фермеры", icon: "🌾" },
    { id: "clients", label: "Клиенты и бонусы", icon: "👥" },
    { id: "recommendations", label: "Рекомендации", icon: "🤖" },
  ];

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-logo">
          <span className="logo-icon">🚜</span>
          <h1>MISIS Kolhoz</h1>
        </div>
        <nav className="sidebar-nav">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              className={`nav-btn ${activeTab === tab.id ? "active" : ""}`}
              onClick={() => setActiveTab(tab.id)}
            >
              <span className="nav-icon">{tab.icon}</span>
              <span>{tab.label}</span>
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className={`status-dot ${healthStatus === true ? "online" : healthStatus === false ? "offline" : "unknown"}`} />
          <span className="status-text">
            {healthStatus === true ? "API онлайн" : healthStatus === false ? "API оффлайн" : "Проверка..."}
          </span>
        </div>
      </aside>

      <main className="main-content">
        <header className="page-header">
          <h2>
            {tabs.find((t) => t.id === activeTab)?.icon}{" "}
            {tabs.find((t) => t.id === activeTab)?.label}
          </h2>
          <p className="header-subtitle">
            {activeTab === "upload" && "Загрузка данных из Excel файлов в базу данных"}
            {activeTab === "farmers" && "Поиск информации о фермере по ID"}
            {activeTab === "clients" && "Управление клиентами и бонусной системой"}
            {activeTab === "recommendations" && "Рекомендации товаров для клиентов на основе истории покупок"}
          </p>
        </header>

        <div className="content-area">
          {activeTab === "upload" && <DataUpload />}
          {activeTab === "farmers" && <FarmerSearch />}
          {activeTab === "clients" && <Clients />}
          {activeTab === "recommendations" && <Recommendations />}
        </div>
      </main>
    </div>
  );
}

export default App;