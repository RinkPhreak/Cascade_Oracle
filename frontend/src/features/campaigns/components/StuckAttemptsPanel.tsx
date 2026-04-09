import { useStuckTasks, useRequeueTask } from '../hooks/useCampaignTasks';
import type { CampaignTask } from '../../../api/extended-types';

interface StuckAttemptsPanelProps {
  campaignId: string;
}

const formatElapsed = (startedAt: string): string => {
  const secs = Math.floor((Date.now() - new Date(startedAt).getTime()) / 1000);
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins} мин`;
  const hours = Math.floor(mins / 60);
  return `${hours} ч ${mins % 60} мин`;
};

const StuckTaskRow = ({ task, onRequeue }: { task: CampaignTask; onRequeue: (id: string) => void }) => (
  <div className="flex items-center gap-4 py-3 border-b border-neutral-800 last:border-0">
    <div className="flex-1 min-w-0">
      <p className="text-sm font-mono text-white truncate">{task.contact_phone}</p>
      <p className="text-xs text-neutral-500 mt-0.5">
        Канал: <span className="text-neutral-300 uppercase">{task.channel}</span>
        {task.error_code && (
          <> · Ошибка: <span className="text-orange-400 font-mono">{task.error_code}</span></>
        )}
      </p>
    </div>
    <div className="text-right shrink-0">
      <p className="text-xs text-orange-400 font-mono font-bold">
        {task.started_at ? formatElapsed(task.started_at) : '?'}
      </p>
      <p className="text-xs text-neutral-600">зависла</p>
    </div>
    <button
      id={`requeue-task-${task.id}`}
      onClick={() => onRequeue(task.id)}
      className="text-xs px-3 py-1.5 bg-accent-cyan/10 hover:bg-accent-cyan/20 border border-accent-cyan/30
        text-accent-cyan rounded-md transition-colors font-mono whitespace-nowrap"
    >
      ↻ Повторить
    </button>
  </div>
);

export const StuckAttemptsPanel = ({ campaignId }: StuckAttemptsPanelProps) => {
  const { data: stuckTasks = [], isLoading } = useStuckTasks(campaignId);
  const requeueTask = useRequeueTask(campaignId);

  if (isLoading) return null;

  if (stuckTasks.length === 0) {
    return (
      <div className="flex items-center gap-2 text-green-400 text-xs bg-green-500/5 border border-green-500/15 rounded-lg px-4 py-3">
        <span>✓</span>
        <span>Зависших задач нет</span>
      </div>
    );
  }

  return (
    <div className="bg-orange-500/5 border border-orange-500/20 rounded-xl overflow-hidden">
      <div className="flex items-center gap-3 px-5 py-4 border-b border-orange-500/20">
        <span className="banner-pulse inline-block w-2.5 h-2.5 rounded-full bg-orange-400 shrink-0" />
        <div>
          <p className="text-sm font-bold text-orange-400">
            Зависшие задачи: {stuckTasks.length}
          </p>
          <p className="text-xs text-neutral-500 mt-0.5">
            Задачи в статусе `in_progress` более 10 минут — требуют вмешательства
          </p>
        </div>
      </div>
      <div className="px-5 divide-y divide-neutral-800">
        {stuckTasks.map((task) => (
          <StuckTaskRow
            key={task.id}
            task={task}
            onRequeue={(id) => requeueTask.mutate(id)}
          />
        ))}
      </div>
    </div>
  );
};
