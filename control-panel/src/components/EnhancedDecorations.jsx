import { motion, useScroll, useTransform } from 'framer-motion';
import { useRef, useEffect, useState, useMemo } from 'react';

// Animated particles floating upward
export function FloatingParticles({ count = 30, className = '' }) {
  const particles = useMemo(() =>
    Array.from({ length: count }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      delay: Math.random() * 5,
      duration: 5 + Math.random() * 5,
      size: 2 + Math.random() * 4,
      opacity: 0.2 + Math.random() * 0.3
    })),
    [count]
  );

  return (
    <div className={`absolute inset-0 overflow-hidden pointer-events-none ${className}`}>
      {particles.map((p) => (
        <motion.div
          key={p.id}
          className="absolute rounded-full bg-gold-500"
          style={{
            left: `${p.x}%`,
            width: p.size,
            height: p.size,
            opacity: p.opacity
          }}
          animate={{
            y: [window.innerHeight, -50],
            x: [0, Math.sin(p.id) * 30, 0],
            opacity: [0, p.opacity, p.opacity, 0]
          }}
          transition={{
            duration: p.duration,
            delay: p.delay,
            repeat: Infinity,
            ease: "linear"
          }}
        />
      ))}
    </div>
  );
}

// Glowing network lines with data packets
export function NetworkLinesEffect({ className = '' }) {
  const lines = [
    { start: [10, 30], end: [90, 20], packets: 3 },
    { start: [5, 50], end: [95, 60], packets: 2 },
    { start: [15, 80], end: [85, 70], packets: 4 },
    { start: [50, 10], end: [50, 90], packets: 2 },
    { start: [20, 20], end: [80, 80], packets: 3 },
    { start: [80, 20], end: [20, 80], packets: 2 },
  ];

  return (
    <svg className={`absolute inset-0 w-full h-full ${className}`} preserveAspectRatio="none">
      <defs>
        <linearGradient id="lineGradient" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#F59E0B" stopOpacity="0" />
          <stop offset="50%" stopColor="#F59E0B" stopOpacity="0.3" />
          <stop offset="100%" stopColor="#F59E0B" stopOpacity="0" />
        </linearGradient>
        <filter id="glow">
          <feGaussianBlur stdDeviation="2" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      {lines.map((line, i) => (
        <g key={i}>
          {/* Base line */}
          <line
            x1={`${line.start[0]}%`}
            y1={`${line.start[1]}%`}
            x2={`${line.end[0]}%`}
            y2={`${line.end[1]}%`}
            stroke="url(#lineGradient)"
            strokeWidth="1"
          />

          {/* Data packets */}
          {Array.from({ length: line.packets }).map((_, j) => (
            <circle
              key={j}
              r="4"
              fill="#F59E0B"
              filter="url(#glow)"
            >
              <animateMotion
                dur={`${3 + j}s`}
                repeatCount="indefinite"
                begin={`${j * 0.8}s`}
                path={`M${line.start[0] * 4},${line.start[1] * 3} L${line.end[0] * 4},${line.end[1] * 3}`}
              />
              <animate
                attributeName="opacity"
                values="0;1;1;0"
                dur={`${3 + j}s`}
                repeatCount="indefinite"
                begin={`${j * 0.8}s`}
              />
            </circle>
          ))}
        </g>
      ))}
    </svg>
  );
}

// Hexagon grid pattern
export function HexagonGrid({ className = '' }) {
  const hexagons = useMemo(() => {
    const result = [];
    const rows = 8;
    const cols = 12;
    for (let row = 0; row < rows; row++) {
      for (let col = 0; col < cols; col++) {
        const offset = row % 2 === 0 ? 0 : 4;
        result.push({
          id: `${row}-${col}`,
          x: col * 8 + offset,
          y: row * 7,
          delay: (row + col) * 0.1
        });
      }
    }
    return result;
  }, []);

  return (
    <svg className={`absolute inset-0 w-full h-full ${className}`} viewBox="0 0 100 60">
      <defs>
        <linearGradient id="hexGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#F59E0B" stopOpacity="0.1" />
          <stop offset="100%" stopColor="#F59E0B" stopOpacity="0.05" />
        </linearGradient>
      </defs>

      {hexagons.map((hex) => (
        <motion.polygon
          key={hex.id}
          points="3,0 6,1.5 6,4.5 3,6 0,4.5 0,1.5"
          transform={`translate(${hex.x}, ${hex.y})`}
          fill="none"
          stroke="url(#hexGrad)"
          strokeWidth="0.1"
          initial={{ opacity: 0, scale: 0 }}
          animate={{ opacity: 0.5, scale: 1 }}
          transition={{
            duration: 1,
            delay: hex.delay,
            repeat: Infinity,
            repeatType: "reverse",
            repeatDelay: 2
          }}
        />
      ))}
    </svg>
  );
}

// Animated gradient orb
export function AnimatedOrb({
  className = '',
  size = 400,
  color1 = '#F59E0B',
  color2 = '#D97706',
  speed = 4
}) {
  return (
    <motion.div
      className={`absolute rounded-full blur-3xl ${className}`}
      style={{
        width: size,
        height: size,
        background: `radial-gradient(circle, ${color1}40 0%, ${color2}10 50%, transparent 70%)`
      }}
      animate={{
        scale: [1, 1.2, 1],
        opacity: [0.3, 0.5, 0.3],
        x: [0, 30, 0],
        y: [0, -20, 0]
      }}
      transition={{
        duration: speed,
        repeat: Infinity,
        ease: "easeInOut"
      }}
    />
  );
}

// Scrolling code rain effect
export function CodeRain({ className = '' }) {
  const columns = 15;
  const chars = '01アウレオVPN暗号化セキュリティ';

  return (
    <div className={`absolute inset-0 overflow-hidden pointer-events-none ${className}`}>
      {Array.from({ length: columns }).map((_, i) => (
        <motion.div
          key={i}
          className="absolute text-gold-500/20 font-mono text-xs whitespace-nowrap"
          style={{
            left: `${(i / columns) * 100}%`,
            writingMode: 'vertical-rl'
          }}
          animate={{
            y: ['-100%', '100%']
          }}
          transition={{
            duration: 10 + Math.random() * 10,
            repeat: Infinity,
            delay: Math.random() * 5,
            ease: 'linear'
          }}
        >
          {Array.from({ length: 30 }).map((_, j) => (
            <span key={j}>
              {chars[Math.floor(Math.random() * chars.length)]}
            </span>
          ))}
        </motion.div>
      ))}
    </div>
  );
}

// Radar sweep effect
export function RadarSweep({ className = '', size = 300 }) {
  return (
    <div
      className={`relative ${className}`}
      style={{ width: size, height: size }}
    >
      {/* Radar circles */}
      {[1, 0.75, 0.5, 0.25].map((scale, i) => (
        <div
          key={i}
          className="absolute inset-0 rounded-full border border-gold-500/20"
          style={{ transform: `scale(${scale})` }}
        />
      ))}

      {/* Crosshairs */}
      <div className="absolute left-1/2 top-0 bottom-0 w-px bg-gold-500/20" />
      <div className="absolute top-1/2 left-0 right-0 h-px bg-gold-500/20" />

      {/* Sweeping line */}
      <motion.div
        className="absolute inset-0 origin-center"
        animate={{ rotate: 360 }}
        transition={{ duration: 4, repeat: Infinity, ease: 'linear' }}
      >
        <div
          className="absolute left-1/2 top-1/2 w-1/2 h-0.5 origin-left"
          style={{
            background: 'linear-gradient(90deg, #F59E0B, transparent)'
          }}
        />
      </motion.div>

      {/* Blips */}
      {[
        { x: 60, y: 30 },
        { x: 20, y: 60 },
        { x: 70, y: 70 },
        { x: 40, y: 45 }
      ].map((blip, i) => (
        <motion.div
          key={i}
          className="absolute w-2 h-2 bg-gold-500 rounded-full"
          style={{ left: `${blip.x}%`, top: `${blip.y}%` }}
          animate={{
            opacity: [0, 1, 0],
            scale: [0.5, 1, 0.5]
          }}
          transition={{
            duration: 2,
            delay: i * 0.5,
            repeat: Infinity
          }}
        />
      ))}
    </div>
  );
}

// Morphing blob
export function MorphingBlob({ className = '', color = '#F59E0B' }) {
  return (
    <motion.div
      className={`${className}`}
      animate={{
        borderRadius: [
          '60% 40% 30% 70% / 60% 30% 70% 40%',
          '30% 60% 70% 40% / 50% 60% 30% 60%',
          '60% 40% 30% 70% / 60% 30% 70% 40%'
        ]
      }}
      transition={{
        duration: 8,
        repeat: Infinity,
        ease: 'easeInOut'
      }}
      style={{
        background: `linear-gradient(135deg, ${color}30, ${color}10)`,
        filter: 'blur(40px)'
      }}
    />
  );
}

// Typing cursor effect
export function TypingText({ texts, className = '' }) {
  const [textIndex, setTextIndex] = useState(0);
  const [charIndex, setCharIndex] = useState(0);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    const currentText = texts[textIndex];

    const timeout = setTimeout(() => {
      if (!isDeleting && charIndex < currentText.length) {
        setCharIndex(charIndex + 1);
      } else if (!isDeleting && charIndex === currentText.length) {
        setTimeout(() => setIsDeleting(true), 1500);
      } else if (isDeleting && charIndex > 0) {
        setCharIndex(charIndex - 1);
      } else if (isDeleting && charIndex === 0) {
        setIsDeleting(false);
        setTextIndex((textIndex + 1) % texts.length);
      }
    }, isDeleting ? 50 : 100);

    return () => clearTimeout(timeout);
  }, [charIndex, isDeleting, textIndex, texts]);

  return (
    <span className={className}>
      {texts[textIndex].substring(0, charIndex)}
      <motion.span
        animate={{ opacity: [1, 0] }}
        transition={{ duration: 0.5, repeat: Infinity }}
        className="inline-block w-0.5 h-[1em] bg-gold-500 ml-1"
      />
    </span>
  );
}

// Spotlight effect that follows scroll
export function SpotlightEffect({ className = '' }) {
  const [mousePosition, setMousePosition] = useState({ x: 50, y: 50 });

  useEffect(() => {
    const handleMouseMove = (e) => {
      setMousePosition({
        x: (e.clientX / window.innerWidth) * 100,
        y: (e.clientY / window.innerHeight) * 100
      });
    };

    window.addEventListener('mousemove', handleMouseMove);
    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, []);

  return (
    <div
      className={`absolute inset-0 pointer-events-none ${className}`}
      style={{
        background: `radial-gradient(circle at ${mousePosition.x}% ${mousePosition.y}%, rgba(245, 158, 11, 0.1) 0%, transparent 50%)`
      }}
    />
  );
}

// Animated border beam
export function BorderBeam({ className = '', duration = 3 }) {
  return (
    <div className={`absolute inset-0 overflow-hidden rounded-inherit ${className}`}>
      <motion.div
        className="absolute w-[200%] h-[200%] -top-1/2 -left-1/2"
        style={{
          background: 'conic-gradient(from 0deg, transparent 0deg, #F59E0B 10deg, transparent 60deg)'
        }}
        animate={{ rotate: 360 }}
        transition={{ duration, repeat: Infinity, ease: 'linear' }}
      />
    </div>
  );
}
