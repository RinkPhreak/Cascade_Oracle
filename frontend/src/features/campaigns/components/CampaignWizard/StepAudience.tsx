import { useCallback, useRef, useState } from 'react';
import Papa from 'papaparse';

export interface ParsedContact {
  phone: string;
  name: string;
  extra_data: Record<string, string>;
}

interface StepAudienceProps {
  file: File | null;
  parsedContacts: ParsedContact[];
  onFileChange: (file: File | null, parsed: ParsedContact[]) => void;
}

/** Parse CSV using PapaParse expecting columns: phone, name, extra_data (JSON or plain). */
function parseCsv(text: string): ParsedContact[] {
  const parsed = Papa.parse(text, {
    header: true,
    skipEmptyLines: true,
  });

  if (!parsed.data || parsed.data.length === 0) return [];

  const results: ParsedContact[] = [];

  for (const row of parsed.data as Record<string, string>[]) {
    // Normalize keys to lowercase for robust matching
    const normalizedRow: Record<string, string> = {};
    for (const key in row) {
      if (Object.prototype.hasOwnProperty.call(row, key)) {
        normalizedRow[key.trim().toLowerCase()] = row[key];
      }
    }

    const phone = normalizedRow['phone']?.trim();
    if (!phone) continue;

    const name = normalizedRow['name']?.trim() ?? '';
    const rawExtra = normalizedRow['extra_data']?.trim() ?? '';

    let extra_data: Record<string, string> = {};
    if (rawExtra) {
      try {
        extra_data = JSON.parse(rawExtra);
      } catch {
        extra_data = { value: rawExtra }; // Fallback to plain string mapping
      }
    }

    results.push({ phone, name, extra_data });
  }

  return results;
}

export const StepAudience = ({ file, parsedContacts, onFileChange }: StepAudienceProps) => {
  const [dragOver, setDragOver] = useState(false);
  const [parseError, setParseError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const processFile = useCallback(
    (f: File) => {
      if (!f.name.endsWith('.csv')) {
        setParseError('Поддерживаются только файлы .csv');
        onFileChange(null, []);
        return;
      }
      const reader = new FileReader();
      reader.onload = (e) => {
        const text = e.target?.result as string;
        try {
          const parsed = parseCsv(text);
          if (parsed.length === 0) {
            setParseError(
              'CSV не содержит валидных строк. Убедитесь, что первая строка содержит заголовки: phone, name, extra_data'
            );
            onFileChange(f, []);
          } else {
            setParseError(null);
            onFileChange(f, parsed);
          }
        } catch {
          setParseError('Не удалось распарсить CSV файл.');
          onFileChange(f, []);
        }
      };
      reader.readAsText(f, 'UTF-8');
    },
    [onFileChange]
  );

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) processFile(dropped);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (selected) processFile(selected);
  };

  return (
    <div className="flex flex-col gap-5">
      {/* Drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        className={`border-2 border-dashed rounded-xl p-8 flex flex-col items-center gap-3
          cursor-pointer transition-all ${dragOver
            ? 'border-accent-cyan bg-accent-cyan/5'
            : file
              ? 'border-green-500/50 bg-green-500/5'
              : 'border-neutral-700 hover:border-neutral-500 hover:bg-neutral-900/50'
          }`}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv"
          id="audience-csv-input"
          className="hidden"
          onChange={handleInputChange}
        />
        <div className="text-3xl">{file ? '✅' : '📄'}</div>
        {file ? (
          <>
            <p className="text-white font-medium text-sm">{file.name}</p>
            <p className="text-neutral-400 text-xs">{(file.size / 1024).toFixed(1)} KB</p>
          </>
        ) : (
          <>
            <p className="text-neutral-300 text-sm font-medium">
              Перетащите CSV файл или <span className="text-accent-cyan">нажмите для выбора</span>
            </p>
            <p className="text-neutral-500 text-xs font-mono">
              Обязательные колонки: <span className="text-neutral-300">phone, name, extra_data</span>
            </p>
          </>
        )}
      </div>

      {parseError && (
        <div className="flex items-start gap-2 text-danger text-sm bg-danger/5 border border-danger/20 rounded-lg px-4 py-3">
          <span className="text-danger shrink-0">⚠</span>
          <span>{parseError}</span>
        </div>
      )}

      {/* Preview table */}
      {parsedContacts.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <p className="text-xs font-mono text-neutral-400">
              Предпросмотр: <span className="text-white">{parsedContacts.length}</span> контактов
            </p>
            <button
              onClick={(e) => { e.stopPropagation(); onFileChange(null, []); setParseError(null); }}
              className="text-xs text-neutral-500 hover:text-danger transition-colors"
            >
              × Очистить
            </button>
          </div>
          <div className="max-h-44 overflow-y-auto rounded-lg border border-neutral-800">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-neutral-900">
                <tr>
                  {['phone', 'name', 'extra_data'].map((h) => (
                    <th key={h} className="px-3 py-2 text-left font-mono text-neutral-500 uppercase">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {parsedContacts.slice(0, 5).map((c, i) => (
                  <tr key={i} className="border-t border-neutral-800">
                    <td className="px-3 py-1.5 font-mono text-neutral-300">{c.phone}</td>
                    <td className="px-3 py-1.5 text-neutral-300">{c.name || <span className="text-neutral-600">—</span>}</td>
                    <td className="px-3 py-1.5 text-neutral-500 truncate max-w-[120px]">
                      {Object.keys(c.extra_data).length > 0
                        ? JSON.stringify(c.extra_data)
                        : <span className="text-neutral-700">—</span>}
                    </td>
                  </tr>
                ))}
                {parsedContacts.length > 5 && (
                  <tr>
                    <td colSpan={3} className="px-3 py-2 text-center text-neutral-600 italic">
                      … и ещё {parsedContacts.length - 5} строк
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
