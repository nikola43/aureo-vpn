import { motion, useScroll, useTransform } from 'framer-motion';
import { useRef } from 'react';

export function FloatingShield({ className = '', delay = 0 }) {
  return (
    <motion.svg
      viewBox="0 0 100 120"
      className={className}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.8 }}
    >
      <defs>
        <linearGradient id="shieldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#FBBF24" stopOpacity="0.8" />
          <stop offset="100%" stopColor="#D97706" stopOpacity="0.4" />
        </linearGradient>
        <filter id="shieldGlow">
          <feGaussianBlur stdDeviation="2" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <motion.path
        d="M50 10 L90 30 L90 70 C90 95 70 110 50 115 C30 110 10 95 10 70 L10 30 L50 10Z"
        fill="url(#shieldGrad)"
        filter="url(#shieldGlow)"
        animate={{
          y: [0, -10, 0],
        }}
        transition={{
          duration: 4,
          repeat: Infinity,
          ease: "easeInOut",
          delay
        }}
      />
      <motion.path
        d="M50 25 L75 40 L75 65 C75 82 62 92 50 95 C38 92 25 82 25 65 L25 40 L50 25Z"
        fill="none"
        stroke="#F59E0B"
        strokeWidth="1"
        strokeOpacity="0.5"
        animate={{
          opacity: [0.3, 0.7, 0.3],
        }}
        transition={{
          duration: 2,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
    </motion.svg>
  );
}

export function DataStream({ className = '' }) {
  const particles = Array.from({ length: 20 }, (_, i) => ({
    id: i,
    delay: Math.random() * 3,
    duration: 2 + Math.random() * 2,
    x: Math.random() * 100,
    size: 2 + Math.random() * 3
  }));

  return (
    <svg className={className} viewBox="0 0 200 400" preserveAspectRatio="none">
      <defs>
        <linearGradient id="streamGrad" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#F59E0B" stopOpacity="0" />
          <stop offset="50%" stopColor="#F59E0B" stopOpacity="0.8" />
          <stop offset="100%" stopColor="#F59E0B" stopOpacity="0" />
        </linearGradient>
      </defs>
      {particles.map((p) => (
        <motion.circle
          key={p.id}
          r={p.size}
          fill="url(#streamGrad)"
          initial={{ cy: -20, cx: p.x }}
          animate={{ cy: 420 }}
          transition={{
            duration: p.duration,
            repeat: Infinity,
            delay: p.delay,
            ease: "linear"
          }}
        />
      ))}
    </svg>
  );
}

export function OrbitRing({ className = '', reverse = false }) {
  return (
    <motion.svg
      viewBox="0 0 300 300"
      className={className}
      animate={{ rotate: reverse ? -360 : 360 }}
      transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
    >
      <defs>
        <linearGradient id="orbitGrad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#F59E0B" stopOpacity="0" />
          <stop offset="50%" stopColor="#F59E0B" stopOpacity="0.5" />
          <stop offset="100%" stopColor="#F59E0B" stopOpacity="0" />
        </linearGradient>
      </defs>
      <circle
        cx="150"
        cy="150"
        r="120"
        fill="none"
        stroke="url(#orbitGrad)"
        strokeWidth="1"
        strokeDasharray="4 4"
      />
      <circle cx="270" cy="150" r="6" fill="#F59E0B">
        <animate
          attributeName="opacity"
          values="1;0.5;1"
          dur="2s"
          repeatCount="indefinite"
        />
      </circle>
    </motion.svg>
  );
}

export function GlowingNode({ className = '', pulseDelay = 0 }) {
  return (
    <motion.div
      className={`relative ${className}`}
      animate={{
        scale: [1, 1.1, 1],
        opacity: [0.7, 1, 0.7]
      }}
      transition={{
        duration: 3,
        repeat: Infinity,
        delay: pulseDelay,
        ease: "easeInOut"
      }}
    >
      <div className="absolute inset-0 bg-gold-500/30 blur-xl rounded-full" />
      <div className="relative w-4 h-4 bg-gradient-to-br from-gold-400 to-gold-600 rounded-full" />
    </motion.div>
  );
}

export function WavePattern({ className = '' }) {
  return (
    <svg className={className} viewBox="0 0 1440 320" preserveAspectRatio="none">
      <motion.path
        fill="rgba(245, 158, 11, 0.1)"
        d="M0,160L48,144C96,128,192,96,288,106.7C384,117,480,171,576,192C672,213,768,203,864,181.3C960,160,1056,128,1152,128C1248,128,1344,160,1392,176L1440,192L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"
        animate={{
          d: [
            "M0,160L48,144C96,128,192,96,288,106.7C384,117,480,171,576,192C672,213,768,203,864,181.3C960,160,1056,128,1152,128C1248,128,1344,160,1392,176L1440,192L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z",
            "M0,192L48,176C96,160,192,128,288,138.7C384,149,480,203,576,213.3C672,224,768,192,864,165.3C960,139,1056,117,1152,128C1248,139,1344,181,1392,202.7L1440,224L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z",
            "M0,160L48,144C96,128,192,96,288,106.7C384,117,480,171,576,192C672,213,768,203,864,181.3C960,160,1056,128,1152,128C1248,128,1344,160,1392,176L1440,192L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"
          ]
        }}
        transition={{
          duration: 10,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
    </svg>
  );
}

export function GridBackground({ className = '' }) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start']
  });

  const y = useTransform(scrollYProgress, [0, 1], ['0%', '20%']);
  const opacity = useTransform(scrollYProgress, [0, 0.5, 1], [0.05, 0.1, 0.05]);

  return (
    <motion.div
      ref={ref}
      style={{ y, opacity }}
      className={`absolute inset-0 ${className}`}
    >
      <svg width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="grid" width="50" height="50" patternUnits="userSpaceOnUse">
            <path
              d="M 50 0 L 0 0 0 50"
              fill="none"
              stroke="rgba(245, 158, 11, 0.3)"
              strokeWidth="0.5"
            />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
      </svg>
    </motion.div>
  );
}

export function ConnectionLines({ className = '' }) {
  return (
    <svg className={className} viewBox="0 0 400 300">
      <defs>
        <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#F59E0B" stopOpacity="0" />
          <stop offset="50%" stopColor="#F59E0B" stopOpacity="0.5" />
          <stop offset="100%" stopColor="#F59E0B" stopOpacity="0" />
        </linearGradient>
      </defs>

      {/* Nodes */}
      {[
        [50, 50], [350, 50], [200, 150], [50, 250], [350, 250]
      ].map(([x, y], i) => (
        <g key={i}>
          <motion.circle
            cx={x}
            cy={y}
            r="8"
            fill="#1f2937"
            stroke="#F59E0B"
            strokeWidth="2"
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 2, repeat: Infinity, delay: i * 0.3 }}
          />
          <motion.circle
            cx={x}
            cy={y}
            r="3"
            fill="#F59E0B"
            animate={{ opacity: [0.5, 1, 0.5] }}
            transition={{ duration: 1.5, repeat: Infinity, delay: i * 0.2 }}
          />
        </g>
      ))}

      {/* Lines */}
      {[
        [50, 50, 200, 150],
        [350, 50, 200, 150],
        [50, 250, 200, 150],
        [350, 250, 200, 150],
        [50, 50, 350, 50],
        [50, 250, 350, 250]
      ].map(([x1, y1, x2, y2], i) => (
        <motion.line
          key={i}
          x1={x1}
          y1={y1}
          x2={x2}
          y2={y2}
          stroke="url(#lineGrad)"
          strokeWidth="1"
          strokeDasharray="4 4"
          animate={{ strokeDashoffset: [0, -20] }}
          transition={{ duration: 2, repeat: Infinity, ease: "linear", delay: i * 0.1 }}
        />
      ))}
    </svg>
  );
}

export function CryptoIcons({ className = '' }) {
  const cryptos = [
    { symbol: 'ETH', color: '#627EEA', x: 20, y: 30, delay: 0 },
    { symbol: 'BTC', color: '#F7931A', x: 70, y: 60, delay: 0.5 },
    { symbol: 'LTC', color: '#345D9D', x: 40, y: 80, delay: 1 }
  ];

  return (
    <svg className={className} viewBox="0 0 100 100">
      {cryptos.map((crypto, i) => (
        <motion.g
          key={crypto.symbol}
          animate={{
            y: [0, -10, 0],
            opacity: [0.6, 1, 0.6]
          }}
          transition={{
            duration: 3,
            repeat: Infinity,
            delay: crypto.delay,
            ease: "easeInOut"
          }}
        >
          <circle
            cx={crypto.x}
            cy={crypto.y}
            r="12"
            fill={crypto.color}
            opacity="0.2"
          />
          <circle
            cx={crypto.x}
            cy={crypto.y}
            r="8"
            fill={crypto.color}
          />
          <text
            x={crypto.x}
            y={crypto.y + 3}
            textAnchor="middle"
            fontSize="5"
            fill="white"
            fontWeight="bold"
          >
            {crypto.symbol.charAt(0)}
          </text>
        </motion.g>
      ))}
    </svg>
  );
}
