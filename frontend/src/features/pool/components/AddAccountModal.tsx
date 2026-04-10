import { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { Modal } from '../../../shared/components/Modal';
import { useImportAccount } from '../hooks/useAccounts';

interface ImportAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ImportAccountModal = ({ isOpen, onClose }: ImportAccountModalProps) => {
  const [files, setFiles] = useState<File[]>([]);
  const [comment, setComment] = useState('');
  const [proxy, setProxy] = useState({
    host: '',
    port: '',
    username: '',
    password: '',
  });

  const importMutation = useImportAccount();

  const onDrop = useCallback((acceptedFiles: File[]) => {
    setFiles(acceptedFiles);
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/zip': ['.zip'],
      'application/x-sqlite3': ['.session'],
      'application/json': ['.json'],
    },
    multiple: true,
  });

  const handleImport = async () => {
    if (files.length === 0) return;
    if (!proxy.host || !proxy.port) return;

    const formData = new FormData();
    files.forEach((file) => {
      formData.append('files', file);
    });
    formData.append('proxy_host', proxy.host);
    formData.append('proxy_port', proxy.port);
    formData.append('proxy_username', proxy.username);
    formData.append('proxy_password', proxy.password);
    formData.append('comment', comment);

    importMutation.mutate(formData, {
      onSuccess: () => {
        setFiles([]);
        setComment('');
        onClose();
      },
    });
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Импорт Telegram аккаунта"
      width="max-w-xl"
    >
      <div className="px-6 py-5 flex flex-col gap-6">
        {/* Dropzone */}
        <div
          {...getRootProps()}
          className={`border-2 border-dashed rounded-xl p-8 transition-colors text-center cursor-pointer
            ${isDragActive ? 'border-accent-cyan bg-accent-cyan/10' : 'border-neutral-700 hover:border-neutral-500 bg-neutral-900/50'}`}
        >
          <input {...getInputProps()} />
          <div className="flex flex-col items-center gap-3">
            <div className="w-12 h-12 rounded-full bg-neutral-800 flex items-center justify-center text-neutral-400">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
            </div>
            {files.length > 0 ? (
              <div className="space-y-1">
                <p className="text-sm text-white font-medium">Выбрано файлов: {files.length}</p>
                <p className="text-xs text-neutral-500">{files.map(f => f.name).join(', ')}</p>
              </div>
            ) : (
              <div className="space-y-1">
                <p className="text-sm text-neutral-300">Перетащите .zip архив или .session + .json файлы</p>
                <p className="text-xs text-neutral-500">Поддерживаются форматы Telethon</p>
              </div>
            )}
          </div>
        </div>

        {/* Proxy Settings */}
        <div className="space-y-4">
          <h3 className="text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
            Настройки Proxy (SOCKS5/MTProto) *
          </h3>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5 col-span-2 sm:col-span-1">
              <label className="text-[10px] text-neutral-500 font-medium ml-1">IP / Хост</label>
              <input
                type="text"
                placeholder="192.168.1.1"
                className="w-full bg-neutral-900 border border-neutral-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-accent-cyan"
                value={proxy.host}
                onChange={(e) => setProxy({ ...proxy, host: e.target.value })}
              />
            </div>
            <div className="space-y-1.5 col-span-2 sm:col-span-1">
              <label className="text-[10px] text-neutral-500 font-medium ml-1">Порт</label>
              <input
                type="number"
                placeholder="1080"
                className="w-full bg-neutral-900 border border-neutral-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-accent-cyan"
                value={proxy.port}
                onChange={(e) => setProxy({ ...proxy, port: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-[10px] text-neutral-500 font-medium ml-1">Логин</label>
              <input
                type="text"
                className="w-full bg-neutral-900 border border-neutral-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-accent-cyan"
                value={proxy.username}
                onChange={(e) => setProxy({ ...proxy, username: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-[10px] text-neutral-500 font-medium ml-1">Пароль</label>
              <input
                type="password"
                className="w-full bg-neutral-900 border border-neutral-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-accent-cyan"
                value={proxy.password}
                onChange={(e) => setProxy({ ...proxy, password: e.target.value })}
              />
            </div>
          </div>
        </div>

        {/* Comment */}
        <div className="space-y-1.5">
          <label className="text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
            Комментарий
          </label>
          <textarea
            placeholder="Название аккаунта или заметка..."
            rows={2}
            className="w-full bg-neutral-900 border border-neutral-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-accent-cyan resize-none"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
          />
        </div>

        {/* Action */}
        <button
          onClick={handleImport}
          disabled={files.length === 0 || !proxy.host || !proxy.port || importMutation.isPending}
          className="w-full bg-accent-cyan text-bg-dark font-bold py-3 rounded-lg hover:bg-cyan-400 transition-colors
            disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {importMutation.isPending && (
            <div className="w-4 h-4 border-2 border-bg-dark/30 border-t-bg-dark rounded-full animate-spin" />
          )}
          Начать импорт
        </button>
      </div>
    </Modal>
  );
};
