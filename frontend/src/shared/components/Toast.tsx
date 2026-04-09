import { useToastStore } from '../hooks/useToast';

const variantStyles: Record<string, string> = {
  success: 'border-green-500 bg-green-500/10 text-green-400',
  error: 'border-danger bg-danger/10 text-red-400',
  warning: 'border-orange-500 bg-orange-500/10 text-orange-400',
  info: 'border-accent-cyan bg-accent-cyan/10 text-accent-cyan',
};

const variantIcons: Record<string, string> = {
  success: '✓',
  error: '✕',
  warning: '⚠',
  info: 'ℹ',
};

export const ToastContainer = () => {
  const { toasts, removeToast } = useToastStore();

  return (
    <div
      aria-live="assertive"
      className="fixed bottom-6 right-6 z-[9999] flex flex-col gap-3 max-w-sm w-full pointer-events-none"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`flex items-start gap-3 px-4 py-3 rounded-lg border backdrop-blur-sm
            shadow-lg pointer-events-auto toast-slide
            ${variantStyles[toast.variant]}`}
          role="alert"
        >
          <span className="text-base font-bold mt-0.5 shrink-0">
            {variantIcons[toast.variant]}
          </span>
          <p className="text-sm flex-1 leading-snug">{toast.message}</p>
          <button
            onClick={() => removeToast(toast.id)}
            className="opacity-60 hover:opacity-100 transition-opacity text-xs ml-1 shrink-0"
            aria-label="Dismiss"
          >
            ✕
          </button>
        </div>
      ))}
    </div>
  );
};
