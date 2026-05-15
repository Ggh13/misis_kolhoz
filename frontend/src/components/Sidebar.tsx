import {
  LayoutDashboard,
  Calendar,
  ShoppingCart,
  Sparkles,
  TrendingUp,
  Package,
  Database,
  User,
  ChevronLeft,
  ChevronRight,
  CheckCircle,
} from 'lucide-react';
import type { FarmerInfo } from '../types';

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  collapsed: boolean;
  onToggleCollapse: () => void;
  selectedFarmer: FarmerInfo | null;
}

export function Sidebar({ activeTab, onTabChange, collapsed, onToggleCollapse, selectedFarmer }: SidebarProps) {
  const menuItems = [
    { id: 'dashboard', icon: LayoutDashboard, label: 'Дашборд' },
    { id: 'ingestion', icon: Database, label: 'Загрузка данных' },
    { id: 'events', icon: Calendar, label: 'События и тренды' },
    { id: 'matcher', icon: Sparkles, label: 'Сопоставление' },
    { id: 'products', icon: Package, label: 'Товары' },
    { id: 'analytics', icon: TrendingUp, label: 'Аналитика' },
  ];

  return (
    <div className={`${collapsed ? 'w-16' : 'w-60'} bg-white border-r border-gray-200 h-screen flex flex-col transition-all duration-300 relative`}>
      <button
        onClick={onToggleCollapse}
        className="absolute -right-3 top-6 w-6 h-6 bg-white border border-gray-200 rounded-full flex items-center justify-center hover:bg-gray-50 transition-colors shadow-sm z-10"
      >
        {collapsed ? <ChevronRight className="w-3 h-3 text-gray-600" /> : <ChevronLeft className="w-3 h-3 text-gray-600" />}
      </button>

      <div className={`px-4 py-5 ${collapsed ? 'items-center' : ''}`}>
        <div className={`flex items-center ${collapsed ? 'justify-center' : 'gap-2.5'}`}>
          <div className="w-8 h-8 bg-green-600 rounded-md flex items-center justify-center flex-shrink-0">
            <ShoppingCart className="w-4 h-4 text-white" />
          </div>
          {!collapsed && (
            <div>
              <h1 className="text-sm font-semibold text-gray-900">AgroTrend AI</h1>
              <p className="text-xs text-gray-500">Событийный маркетинг</p>
            </div>
          )}
        </div>
      </div>

      <nav className="flex-1 px-3 pb-3 overflow-y-auto">
        {menuItems.map((item) => {
          const Icon = item.icon;
          return (
            <button
              key={item.id}
              data-tab={item.id}
              onClick={() => onTabChange(item.id)}
              title={collapsed ? item.label : undefined}
              className={`w-full flex items-center ${collapsed ? 'justify-center' : 'gap-2.5'} px-3 py-2 rounded-md mb-0.5 transition-colors text-sm ${
                activeTab === item.id ? 'bg-green-50 text-green-700' : 'text-gray-700 hover:bg-gray-50'
              }`}
            >
              <Icon className="w-4 h-4 flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </button>
          );
        })}
      </nav>

      <div className="p-4 border-t border-gray-200">
        <div className={`flex items-center ${collapsed ? 'justify-center' : 'gap-2.5'}`}>
          <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${selectedFarmer ? 'bg-green-100' : 'bg-gray-200'}`}>
            {selectedFarmer ? <CheckCircle className="w-4 h-4 text-green-600" /> : <User className="w-4 h-4 text-gray-600" />}
          </div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              {selectedFarmer ? (
                <>
                  <p className="text-sm font-medium text-gray-900 truncate">{selectedFarmer.name || `Фермер #${selectedFarmer.id}`}</p>
                  <p className="text-xs text-green-600 truncate">Выбран</p>
                </>
              ) : (
                <>
                  <p className="text-sm font-medium text-gray-900 truncate">Не выбран</p>
                  <p className="text-xs text-gray-500 truncate">Выберите в загрузке</p>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}