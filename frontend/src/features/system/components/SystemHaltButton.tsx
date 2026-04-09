export const SystemHaltButton = () => {
  const handleHaltClick = () => {
    console.log("Halt click requested: Waiting for modal wireframe implementation.");
  };

  return (
    <button 
      onClick={handleHaltClick}
      className="bg-danger/10 hover:bg-danger/20 border border-danger text-danger px-4 py-2 rounded-md font-bold text-sm transition-colors uppercase tracking-wider"
    >
      Break Glass Halt
    </button>
  );
};
