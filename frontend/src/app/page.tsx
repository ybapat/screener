import Link from "next/link";

export default function Home() {
  return (
    <main className="min-h-screen flex flex-col">
      {/* Nav */}
      <nav className="flex items-center justify-between px-8 py-5 border-b border-white/10">
        <span className="text-xl font-bold tracking-tight">Screener</span>
        <div className="flex items-center gap-4">
          <Link
            href="/login"
            className="text-sm text-zinc-400 hover:text-white transition-colors"
          >
            Log in
          </Link>
          <Link
            href="/register"
            className="text-sm bg-white text-black px-4 py-2 rounded-lg font-medium hover:bg-zinc-200 transition-colors"
          >
            Get started
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="flex-1 flex flex-col items-center justify-center px-6 text-center">
        <div className="max-w-3xl mx-auto">
          <h1 className="text-6xl sm:text-7xl font-bold tracking-tight mb-6">
            Screener
          </h1>
          <p className="text-xl sm:text-2xl text-zinc-400 mb-12 max-w-xl mx-auto">
            Your screen time has value. Sell it anonymously.
          </p>

          {/* Steps */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-8 mb-14">
            <div className="bg-white/5 border border-white/10 rounded-2xl p-6">
              <div className="w-10 h-10 bg-emerald-500/20 text-emerald-400 rounded-lg flex items-center justify-center font-bold text-lg mb-4 mx-auto">
                1
              </div>
              <h3 className="font-semibold mb-2">Upload your data</h3>
              <p className="text-sm text-zinc-500">
                Share your screen time records securely. You control what gets
                shared.
              </p>
            </div>
            <div className="bg-white/5 border border-white/10 rounded-2xl p-6">
              <div className="w-10 h-10 bg-blue-500/20 text-blue-400 rounded-lg flex items-center justify-center font-bold text-lg mb-4 mx-auto">
                2
              </div>
              <h3 className="font-semibold mb-2">We anonymize it</h3>
              <p className="text-sm text-zinc-500">
                Differential privacy and k-anonymity ensure your identity is
                never exposed.
              </p>
            </div>
            <div className="bg-white/5 border border-white/10 rounded-2xl p-6">
              <div className="w-10 h-10 bg-purple-500/20 text-purple-400 rounded-lg flex items-center justify-center font-bold text-lg mb-4 mx-auto">
                3
              </div>
              <h3 className="font-semibold mb-2">Earn credits</h3>
              <p className="text-sm text-zinc-500">
                When buyers purchase anonymized datasets, you earn credits
                automatically.
              </p>
            </div>
          </div>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/register?role=seller"
              className="w-full sm:w-auto bg-emerald-500 hover:bg-emerald-400 text-black font-semibold px-8 py-3 rounded-lg transition-colors text-center"
            >
              Start Selling
            </Link>
            <Link
              href="/marketplace"
              className="w-full sm:w-auto bg-white/10 hover:bg-white/20 text-white font-semibold px-8 py-3 rounded-lg transition-colors border border-white/10 text-center"
            >
              Browse Marketplace
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-6 px-8 border-t border-white/10 text-center text-sm text-zinc-600">
        Screener -- Privacy-preserving screen time data marketplace
      </footer>
    </main>
  );
}
