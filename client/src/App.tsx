import './App.css';

const AUDIO_BAR_HEIGHTS = [24, 32, 20, 40, 28, 36, 22, 30];
const AUDIO_BAR_OPACITIES = [0.3, 0.5, 0.25, 0.45, 0.35, 0.4, 0.3, 0.5];

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-800 to-slate-900 flex items-center justify-center p-4">
      <div className="w-full max-w-sm bg-gradient-to-b from-red-600 to-red-700 rounded-3xl shadow-2xl border-8 border-slate-800">
        <div className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 bg-blue-400 rounded-full border-4 border-blue-600 shadow-inner led-power" />
              <div className="flex gap-2">
                <div className="w-3 h-3 bg-red-500 rounded-full border-2 border-red-700" />
                <div className="w-3 h-3 bg-yellow-400 rounded-full border-2 border-yellow-600" />
                <div className="w-3 h-3 bg-green-400 rounded-full border-2 border-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-slate-800 rounded-lg p-4 border-4 border-slate-900 shadow-inner">
            <div className="bg-gradient-to-b from-blue-900 to-blue-950 rounded p-6 border-4 border-blue-950 min-h-[280px] flex flex-col items-center justify-center space-y-6 screen-content">
              <div className="w-16 h-16 border-8 border-red-600 rounded-lg transform rotate-45 relative">
                <div className="absolute inset-2 bg-red-600 rounded-sm" />
              </div>

              <h1 className="text-green-400 text-xl font-bold tracking-wider text-center font-mono">
                CLEAN ARCH
                <br />
                POKÉDEX
              </h1>

              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                <p className="text-green-400 text-lg font-mono loading-text">Loading</p>
              </div>

              <div className="flex gap-1">
                {AUDIO_BAR_HEIGHTS.map((height, i) => (
                  <div
                    key={i}
                    className="w-1 bg-green-900 border border-green-700"
                    style={{
                      opacity: AUDIO_BAR_OPACITIES[i],
                      height: `${height.toString()}px`,
                    }}
                  />
                ))}
              </div>
            </div>
          </div>

          <div className="flex items-center justify-between px-2">
            <div className="flex gap-1">
              <div className="w-12 h-12 bg-slate-900 rounded-full border-4 border-slate-950 flex items-center justify-center shadow-inner">
                <div className="w-8 h-8 bg-gradient-to-br from-slate-800 to-slate-900 rounded-full" />
              </div>
            </div>

            <div className="flex gap-3">
              <div className="w-16 h-6 bg-green-600 rounded-full border-2 border-green-800 shadow-md" />
              <div className="w-16 h-6 bg-green-600 rounded-full border-2 border-green-800 shadow-md" />
            </div>
          </div>

          <div className="flex justify-center gap-2 pt-2">
            <div className="w-8 h-8 bg-yellow-500 rounded border-2 border-yellow-700 shadow-md" />
            <div className="w-8 h-8 bg-yellow-500 rounded border-2 border-yellow-700 shadow-md" />
            <div className="w-8 h-8 bg-blue-500 rounded-full border-2 border-blue-700 shadow-md" />
            <div className="w-8 h-8 bg-blue-500 rounded-full border-2 border-blue-700 shadow-md" />
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
