import { Sparkles, X } from 'lucide-react';
import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

export function AIAssistant() {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <AnimatePresence>
        {isOpen && (
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.9, opacity: 0 }} className="fixed bottom-6 right-6 z-50">
            <div className="bg-white rounded-2xl shadow-2xl border border-gray-200 p-5 w-80">
              <div className="flex items-start justify-between mb-3"><h3 className="font-bold text-gray-900 text-sm">AI Ассистент</h3><button onClick={() => setIsOpen(false)}><X className="w-4 h-4 text-gray-400" /></button></div>
              <p className="text-sm text-gray-600">Mock-помощник: «Органическое мясо» показывает экстремальный рост.</p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      {!isOpen && (
        <button onClick={() => setIsOpen(true)} className="fixed bottom-6 right-6 w-14 h-14 bg-gradient-to-br from-green-600 to-emerald-600 rounded-full shadow-2xl flex items-center justify-center z-50">
          <Sparkles className="w-6 h-6 text-white" />
        </button>
      )}
    </>
  );
}
