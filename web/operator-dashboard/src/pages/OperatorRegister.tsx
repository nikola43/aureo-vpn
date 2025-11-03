import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { Wallet, Mail, MapPin, Globe, Sparkles, Zap, CheckCircle } from 'lucide-react';

export const OperatorRegister: React.FC = () => {
  const [walletAddress, setWalletAddress] = useState('');
  const [walletType, setWalletType] = useState<'ethereum' | 'bitcoin' | 'litecoin'>('ethereum');
  const [country, setCountry] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await api.registerAsOperator({
        wallet_address: walletAddress,
        wallet_type: walletType,
        country,
        email,
      });

      // Success! Navigate to dashboard
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen relative overflow-hidden flex items-center justify-center p-4">
      {/* Animated gradient background */}
      <div className="absolute inset-0 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600">
        <div className="absolute inset-0 opacity-30">
          <div className="absolute top-0 left-0 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob"></div>
          <div className="absolute top-0 right-0 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-2000"></div>
          <div className="absolute bottom-0 left-1/2 w-96 h-96 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-4000"></div>
        </div>
      </div>

      {/* Floating particles */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-2 h-2 bg-white/40 rounded-full animate-float"></div>
        <div className="absolute top-1/3 right-1/4 w-3 h-3 bg-white/30 rounded-full animate-float animation-delay-2000"></div>
        <div className="absolute bottom-1/4 left-1/3 w-2 h-2 bg-white/40 rounded-full animate-float animation-delay-4000"></div>
      </div>

      {/* Registration card */}
      <div className="relative w-full max-w-2xl">
        {/* Main card with glass effect */}
        <div className="backdrop-blur-xl bg-white/10 border border-white/20 rounded-3xl shadow-2xl overflow-hidden">
          {/* Gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>

          <div className="relative p-8">
            {/* Header */}
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 mb-4 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 shadow-lg">
                <Sparkles className="h-10 w-10 text-white" />
              </div>
              <h1 className="text-4xl font-bold text-white mb-2 tracking-tight">
                Become an Operator
              </h1>
              <p className="text-white/80 text-lg font-light">Register your node and start earning crypto rewards</p>
            </div>

            {/* Benefits */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
              <div className="backdrop-blur-lg bg-white/5 border border-white/10 rounded-xl p-4 text-center">
                <div className="bg-gradient-to-br from-green-400 to-green-600 p-3 rounded-xl inline-flex mb-2">
                  <CheckCircle className="h-6 w-6 text-white" />
                </div>
                <p className="text-white text-sm font-semibold">Earn Crypto</p>
                <p className="text-white/60 text-xs">Get paid in ETH, BTC, or LTC</p>
              </div>
              <div className="backdrop-blur-lg bg-white/5 border border-white/10 rounded-xl p-4 text-center">
                <div className="bg-gradient-to-br from-purple-400 to-purple-600 p-3 rounded-xl inline-flex mb-2">
                  <Zap className="h-6 w-6 text-white" />
                </div>
                <p className="text-white text-sm font-semibold">Tiered Rewards</p>
                <p className="text-white/60 text-xs">Up to $0.03/GB served</p>
              </div>
              <div className="backdrop-blur-lg bg-white/5 border border-white/10 rounded-xl p-4 text-center">
                <div className="bg-gradient-to-br from-blue-400 to-blue-600 p-3 rounded-xl inline-flex mb-2">
                  <Globe className="h-6 w-6 text-white" />
                </div>
                <p className="text-white text-sm font-semibold">Global Network</p>
                <p className="text-white/60 text-xs">Join operators worldwide</p>
              </div>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-5">
              {/* Wallet Type */}
              <div className="group">
                <label className="block text-sm font-medium text-white/90 mb-2 ml-1">
                  Cryptocurrency
                </label>
                <div className="grid grid-cols-3 gap-3">
                  <button
                    type="button"
                    onClick={() => setWalletType('ethereum')}
                    className={`py-3 px-4 rounded-xl font-semibold transition-all duration-300 ${
                      walletType === 'ethereum'
                        ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg scale-105'
                        : 'backdrop-blur-lg bg-white/5 border border-white/20 text-white/70 hover:bg-white/10'
                    }`}
                  >
                    Ethereum
                  </button>
                  <button
                    type="button"
                    onClick={() => setWalletType('bitcoin')}
                    className={`py-3 px-4 rounded-xl font-semibold transition-all duration-300 ${
                      walletType === 'bitcoin'
                        ? 'bg-gradient-to-r from-orange-500 to-yellow-600 text-white shadow-lg scale-105'
                        : 'backdrop-blur-lg bg-white/5 border border-white/20 text-white/70 hover:bg-white/10'
                    }`}
                  >
                    Bitcoin
                  </button>
                  <button
                    type="button"
                    onClick={() => setWalletType('litecoin')}
                    className={`py-3 px-4 rounded-xl font-semibold transition-all duration-300 ${
                      walletType === 'litecoin'
                        ? 'bg-gradient-to-r from-gray-400 to-gray-600 text-white shadow-lg scale-105'
                        : 'backdrop-blur-lg bg-white/5 border border-white/20 text-white/70 hover:bg-white/10'
                    }`}
                  >
                    Litecoin
                  </button>
                </div>
              </div>

              {/* Wallet Address */}
              <div className="group">
                <label className="block text-sm font-medium text-white/90 mb-2 ml-1">
                  Wallet Address
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <Wallet className="h-5 w-5 text-white/50 group-focus-within:text-white/80 transition-colors" />
                  </div>
                  <input
                    type="text"
                    value={walletAddress}
                    onChange={(e) => setWalletAddress(e.target.value)}
                    className="block w-full pl-12 pr-4 py-3.5 bg-white/10 backdrop-blur-xl border border-white/20 rounded-xl text-white placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent transition-all duration-300 hover:bg-white/15"
                    placeholder={walletType === 'ethereum' ? '0x...' : walletType === 'bitcoin' ? 'bc1...' : 'ltc1...'}
                    required
                  />
                </div>
              </div>

              {/* Email */}
              <div className="group">
                <label className="block text-sm font-medium text-white/90 mb-2 ml-1">
                  Contact Email
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <Mail className="h-5 w-5 text-white/50 group-focus-within:text-white/80 transition-colors" />
                  </div>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="block w-full pl-12 pr-4 py-3.5 bg-white/10 backdrop-blur-xl border border-white/20 rounded-xl text-white placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent transition-all duration-300 hover:bg-white/15"
                    placeholder="your@email.com"
                    required
                  />
                </div>
              </div>

              {/* Country */}
              <div className="group">
                <label className="block text-sm font-medium text-white/90 mb-2 ml-1">
                  Country
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <MapPin className="h-5 w-5 text-white/50 group-focus-within:text-white/80 transition-colors" />
                  </div>
                  <input
                    type="text"
                    value={country}
                    onChange={(e) => setCountry(e.target.value)}
                    className="block w-full pl-12 pr-4 py-3.5 bg-white/10 backdrop-blur-xl border border-white/20 rounded-xl text-white placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent transition-all duration-300 hover:bg-white/15"
                    placeholder="United States"
                    required
                  />
                </div>
              </div>

              {error && (
                <div className="backdrop-blur-xl bg-red-500/20 border border-red-500/30 text-white px-4 py-3.5 rounded-xl text-sm animate-shake">
                  <div className="flex items-center">
                    <Zap className="h-4 w-4 mr-2" />
                    {error}
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="group relative w-full bg-gradient-to-r from-blue-500 via-purple-600 to-pink-600 hover:from-blue-600 hover:via-purple-700 hover:to-pink-700 text-white font-bold py-4 px-6 rounded-xl transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed shadow-xl hover:shadow-2xl hover:scale-[1.02] active:scale-[0.98]"
              >
                <span className="relative z-10 flex items-center justify-center">
                  {loading ? (
                    <>
                      <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Registering...
                    </>
                  ) : (
                    <>
                      Register as Operator
                      <Sparkles className="ml-2 h-5 w-5 group-hover:rotate-12 transition-transform" />
                    </>
                  )}
                </span>
                <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/10 to-white/0 rounded-xl translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000"></div>
              </button>
            </form>
          </div>
        </div>

        {/* Decorative elements */}
        <div className="absolute -top-4 -left-4 w-24 h-24 bg-blue-500/30 rounded-full blur-2xl"></div>
        <div className="absolute -bottom-4 -right-4 w-32 h-32 bg-purple-500/30 rounded-full blur-2xl"></div>
      </div>
    </div>
  );
};
