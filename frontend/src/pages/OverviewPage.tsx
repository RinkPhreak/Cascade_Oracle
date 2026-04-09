export const OverviewPage = () => {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Infrastructure Overview</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-bg-base p-6 rounded-lg border border-neutral-800">
          <div className="text-sm text-neutral-400 mb-2">Active MTProto Sessions</div>
          <div className="text-3xl font-mono text-accent-cyan">0</div>
        </div>
        <div className="bg-bg-base p-6 rounded-lg border border-neutral-800">
          <div className="text-sm text-neutral-400 mb-2">Messages Enqueued</div>
          <div className="text-3xl font-mono text-accent-cyan">0</div>
        </div>
        <div className="bg-bg-base p-6 rounded-lg border border-neutral-800">
          <div className="text-sm text-neutral-400 mb-2">System Memory</div>
          <div className="text-3xl font-mono text-green-500">Normal</div>
        </div>
      </div>
    </div>
  );
};
