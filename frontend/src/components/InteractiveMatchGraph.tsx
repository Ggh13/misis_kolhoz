import { motion } from 'framer-motion';
import { Sparkles } from 'lucide-react';

export function InteractiveMatchGraph() {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-6 h-72 relative overflow-hidden">
      <div className="flex items-center gap-2 mb-4">
        <Sparkles className="w-5 h-5 text-purple-600" />
        <h3 className="text-base font-semibold text-gray-900">Интерактивный граф сопоставления</h3>
      </div>
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-sm text-gray-600">
        Mock-граф: Пасха ↔ Яйца 98%, Творог 92%, Молоко 89%.
      </motion.div>
    </div>
  );
}
