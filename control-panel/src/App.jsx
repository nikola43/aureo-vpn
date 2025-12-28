import React, { useRef, useMemo, Suspense, useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Canvas, useFrame } from '@react-three/fiber';
import { Points, PointMaterial, OrbitControls, Sphere, MeshDistortMaterial, Float, Stars } from '@react-three/drei';
import { motion, useScroll, useTransform, useInView, AnimatePresence, useSpring } from 'framer-motion';
import { useInView as useInViewObserver } from 'react-intersection-observer';
import CountUp from 'react-countup';
import { AreaChart, Area, XAxis, YAxis, ResponsiveContainer, Tooltip } from 'recharts';
import * as THREE from 'three';
import { ParallaxSection, ParallaxLayer, MouseParallax } from './components/ParallaxSection';
import ParallaxImage from './components/ParallaxImage';
import {
  FloatingShield,
  DataStream,
  OrbitRing,
  GlowingNode,
  WavePattern,
  GridBackground,
  ConnectionLines,
  CryptoIcons
} from './components/Decorations';
import {
  ScrollReveal,
  StaggeredContainer,
  StaggeredItem,
  FloatingElement,
  TiltCard,
  MagneticElement,
  ZoomOnScroll,
  RotateOnScroll
} from './components/ScrollAnimations';
import {
  FloatingParticles,
  NetworkLinesEffect,
  HexagonGrid,
  RadarSweep,
  MorphingBlob,
  TypingText
} from './components/EnhancedDecorations';
import NetworkVisualization3D from './components/NetworkVisualization3D';
import {
  Shield,
  Coins,
  Globe,
  Server,
  Activity,
  Lock,
  ArrowRight,
  CheckCircle2,
  Zap,
  Eye,
  EyeOff,
  Cpu,
  Network,
  Wallet,
  TrendingUp,
  Users,
  MapPin,
  ChevronDown,
  Play,
  Star,
  Quote,
  Check,
  X,
  Menu,
  ExternalLink,
  Github,
  Twitter,
  MessageCircle,
  Mail,
  ArrowUpRight,
  Sparkles,
  Bitcoin,
  CircleDollarSign
} from 'lucide-react';

// ============================================
// 3D COMPONENTS
// ============================================

const PARTICLE_COUNT = 5000;

function ParticleField() {
  const ref = useRef();
  const positions = useMemo(() => {
    const pos = new Float32Array(PARTICLE_COUNT * 3);
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      const r = 2 + Math.random() * 0.5;
      pos[i * 3] = r * Math.sin(phi) * Math.cos(theta);
      pos[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
      pos[i * 3 + 2] = r * Math.cos(phi);
    }
    return pos;
  }, []);

  useFrame((state) => {
    if (ref.current) {
      ref.current.rotation.y += 0.0005;
      ref.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.1) * 0.1;
    }
  });

  return (
    <Points ref={ref} positions={positions} stride={3} frustumCulled={false}>
      <PointMaterial
        transparent
        color="#F59E0B"
        size={0.015}
        sizeAttenuation={true}
        depthWrite={false}
        opacity={0.6}
      />
    </Points>
  );
}

function GlobeCore() {
  const meshRef = useRef();

  useFrame((state) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += 0.002;
      meshRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.2) * 0.1;
    }
  });

  return (
    <Sphere ref={meshRef} args={[1.5, 64, 64]}>
      <MeshDistortMaterial
        color="#1f2937"
        attach="material"
        distort={0.3}
        speed={2}
        roughness={0.4}
        metalness={0.8}
      />
    </Sphere>
  );
}

function NetworkNodes() {
  const groupRef = useRef();
  const nodes = useMemo(() => {
    const n = [];
    for (let i = 0; i < 30; i++) {
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      const r = 1.6 + Math.random() * 0.2;
      n.push({
        position: [
          r * Math.sin(phi) * Math.cos(theta),
          r * Math.sin(phi) * Math.sin(theta),
          r * Math.cos(phi)
        ],
        scale: 0.03 + Math.random() * 0.02
      });
    }
    return n;
  }, []);

  useFrame((state) => {
    if (groupRef.current) {
      groupRef.current.rotation.y += 0.001;
    }
  });

  return (
    <group ref={groupRef}>
      {nodes.map((node, i) => (
        <mesh key={i} position={node.position}>
          <sphereGeometry args={[node.scale, 8, 8]} />
          <meshBasicMaterial color="#F59E0B" />
        </mesh>
      ))}
    </group>
  );
}

function Scene3D() {
  return (
    <Canvas
      camera={{ position: [0, 0, 5], fov: 45 }}
      style={{ background: 'transparent' }}
      dpr={[1, 2]}
    >
      <ambientLight intensity={0.5} />
      <pointLight position={[10, 10, 10]} intensity={1} />
      <Suspense fallback={null}>
        <GlobeCore />
        <ParticleField />
        <NetworkNodes />
      </Suspense>
      <OrbitControls
        enableZoom={false}
        enablePan={false}
        autoRotate
        autoRotateSpeed={0.5}
        maxPolarAngle={Math.PI / 2}
        minPolarAngle={Math.PI / 2}
      />
    </Canvas>
  );
}

// ============================================
// ANIMATION VARIANTS
// ============================================

const fadeInUp = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: "easeOut" } }
};

const fadeInLeft = {
  hidden: { opacity: 0, x: -40 },
  visible: { opacity: 1, x: 0, transition: { duration: 0.6, ease: "easeOut" } }
};

const fadeInRight = {
  hidden: { opacity: 0, x: 40 },
  visible: { opacity: 1, x: 0, transition: { duration: 0.6, ease: "easeOut" } }
};

const staggerContainer = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.1, delayChildren: 0.2 }
  }
};

const scaleIn = {
  hidden: { opacity: 0, scale: 0.8 },
  visible: { opacity: 1, scale: 1, transition: { duration: 0.5, ease: "easeOut" } }
};

// ============================================
// REUSABLE COMPONENTS
// ============================================

function AnimatedSection({ children, className = "", delay = 0 }) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });

  return (
    <motion.div
      ref={ref}
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
      variants={{
        hidden: { opacity: 0, y: 50 },
        visible: { opacity: 1, y: 0, transition: { duration: 0.7, delay, ease: "easeOut" } }
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

function GlowingOrb({ className = "", color = "gold" }) {
  const colors = {
    gold: "from-gold-500/30 to-gold-600/10",
    blue: "from-blue-500/20 to-blue-600/5",
    purple: "from-purple-500/20 to-purple-600/5"
  };

  return (
    <div className={`absolute rounded-full blur-3xl animate-pulse-glow ${className}`}>
      <div className={`w-full h-full bg-gradient-radial ${colors[color]}`} />
    </div>
  );
}

function NetworkDiagram() {
  return (
    <svg className="w-full h-full" viewBox="0 0 400 300" fill="none">
      {/* Central Node */}
      <circle cx="200" cy="150" r="30" fill="url(#centerGradient)" className="animate-pulse" />
      <circle cx="200" cy="150" r="40" stroke="#F59E0B" strokeWidth="1" opacity="0.3" className="animate-ping-slow" />

      {/* Outer Nodes */}
      {[0, 60, 120, 180, 240, 300].map((angle, i) => {
        const x = 200 + 100 * Math.cos((angle * Math.PI) / 180);
        const y = 150 + 80 * Math.sin((angle * Math.PI) / 180);
        return (
          <g key={i}>
            <line x1="200" y1="150" x2={x} y2={y} stroke="#F59E0B" strokeWidth="1" opacity="0.3" className="network-line" />
            <circle cx={x} cy={y} r="12" fill="#1f2937" stroke="#F59E0B" strokeWidth="2" />
            <circle cx={x} cy={y} r="4" fill="#F59E0B" className="animate-pulse" />
          </g>
        );
      })}

      {/* Data Flow Particles */}
      {[0, 120, 240].map((angle, i) => {
        const endX = 200 + 100 * Math.cos((angle * Math.PI) / 180);
        const endY = 150 + 80 * Math.sin((angle * Math.PI) / 180);
        return (
          <circle key={`particle-${i}`} r="3" fill="#F59E0B" opacity="0.8">
            <animateMotion
              dur={`${2 + i * 0.5}s`}
              repeatCount="indefinite"
              path={`M200,150 L${endX},${endY}`}
            />
          </circle>
        );
      })}

      <defs>
        <radialGradient id="centerGradient" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#F59E0B" />
          <stop offset="100%" stopColor="#B45309" />
        </radialGradient>
      </defs>
    </svg>
  );
}

// ============================================
// MAIN APP COMPONENT
// ============================================

function App() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const { scrollYProgress } = useScroll();

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <div className="min-h-screen bg-dark-950 text-white overflow-x-hidden">
      {/* Progress Bar */}
      <motion.div
        className="fixed top-0 left-0 right-0 h-1 bg-gradient-to-r from-gold-500 via-gold-400 to-gold-600 z-[100] origin-left"
        style={{ scaleX: scrollYProgress }}
      />

      {/* Navigation */}
      <Navigation scrolled={scrolled} mobileMenuOpen={mobileMenuOpen} setMobileMenuOpen={setMobileMenuOpen} />

      {/* Hero Section */}
      <HeroSection />

      {/* Stats Section */}
      <StatsSection />

      {/* Features Section */}
      <FeaturesSection />

      {/* How it Works */}
      <HowItWorksSection />

      {/* Network Visualization */}
      <NetworkSection />

      {/* Earnings Calculator */}
      <EarningsSection />

      {/* Pricing Section */}
      <PricingSection />

      {/* Testimonials */}
      <TestimonialsSection />

      {/* CTA Section */}
      <CTASection />

      {/* Footer */}
      <Footer />
    </div>
  );
}

// ============================================
// NAVIGATION
// ============================================

function Navigation({ scrolled, mobileMenuOpen, setMobileMenuOpen }) {
  return (
    <nav className={`fixed w-full z-50 transition-all duration-300 ${scrolled ? 'bg-dark-950/90 backdrop-blur-xl border-b border-white/5 py-3' : 'bg-transparent py-5'}`}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <motion.a
            href="#"
            className="flex items-center gap-3 group"
            whileHover={{ scale: 1.02 }}
          >
            <div className="relative">
              <div className="absolute inset-0 bg-gold-500/20 blur-xl rounded-full group-hover:bg-gold-500/30 transition-all" />
              <Shield className="h-9 w-9 text-gold-500 relative" />
            </div>
            <span className="font-bold text-2xl tracking-tight">
              <span className="text-white">Aureo</span>
              <span className="text-gold-500">VPN</span>
            </span>
          </motion.a>

          {/* Desktop Navigation */}
          <div className="hidden lg:flex items-center gap-1">
            {['Features', 'How it Works', 'Earnings', 'Pricing'].map((item) => (
              <a
                key={item}
                href={`#${item.toLowerCase().replace(/\s+/g, '-')}`}
                className="px-4 py-2 text-gray-300 hover:text-white hover:bg-white/5 rounded-lg transition-all font-medium"
              >
                {item}
              </a>
            ))}
          </div>

          {/* CTA Buttons */}
          <div className="hidden lg:flex items-center gap-3">
            <a href="/login" className="px-5 py-2.5 text-gray-300 hover:text-white font-medium transition-all">
              Sign In
            </a>
            <motion.a
              href="/register"
              className="relative group px-6 py-2.5 bg-gradient-to-r from-gold-500 to-gold-600 text-dark-950 font-bold rounded-full overflow-hidden"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.98 }}
            >
              <span className="relative z-10 flex items-center gap-2">
                Get Started
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </span>
              <div className="absolute inset-0 bg-gradient-to-r from-gold-400 to-gold-500 opacity-0 group-hover:opacity-100 transition-opacity" />
            </motion.a>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="lg:hidden p-2 text-gray-400 hover:text-white"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            <Menu className="w-6 h-6" />
          </button>
        </div>
      </div>

      {/* Mobile Menu */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="lg:hidden absolute top-full left-0 right-0 bg-dark-950/95 backdrop-blur-xl border-b border-white/5 p-4"
          >
            <div className="flex flex-col gap-2">
              {['Features', 'How it Works', 'Earnings', 'Pricing'].map((item) => (
                <a
                  key={item}
                  href={`#${item.toLowerCase().replace(/\s+/g, '-')}`}
                  className="px-4 py-3 text-gray-300 hover:text-white hover:bg-white/5 rounded-lg transition-all font-medium"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {item}
                </a>
              ))}
              <hr className="border-white/10 my-2" />
              <a href="/login" className="px-4 py-3 text-gray-300 hover:text-white font-medium">Sign In</a>
              <a href="/register" className="px-4 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg text-center">Get Started</a>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  );
}

// ============================================
// HERO SECTION
// ============================================

function HeroSection() {
  const { scrollYProgress } = useScroll();
  const y = useTransform(scrollYProgress, [0, 1], [0, 300]);
  const opacity = useTransform(scrollYProgress, [0, 0.3], [1, 0]);
  const scale = useTransform(scrollYProgress, [0, 0.5], [1, 0.8]);
  const rotate = useTransform(scrollYProgress, [0, 1], [0, 10]);

  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-20">
      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden">
        <div className="absolute -top-20 -left-20 w-[600px] h-[600px] opacity-20">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=800&auto=format&fit=crop"
            alt="Cyber network"
            className="w-full h-full object-cover rounded-full blur-sm"
            scale={1.5}
            orientation="down"
          />
        </div>
        <div className="absolute -bottom-20 -right-20 w-[500px] h-[500px] opacity-15">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=800&auto=format&fit=crop"
            alt="Digital earth"
            className="w-full h-full object-cover rounded-full blur-sm"
            scale={1.4}
            orientation="up"
          />
        </div>
        <div className="absolute top-1/3 right-10 w-[300px] h-[300px] opacity-10 hidden xl:block">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=600&auto=format&fit=crop"
            alt="Server room"
            className="w-full h-full object-cover rounded-2xl"
            scale={1.3}
            orientation="left"
          />
        </div>
      </div>

      {/* 3D Background with enhanced effects */}
      <motion.div className="absolute inset-0 z-[2]" style={{ y, scale }}>
        <Scene3D />
      </motion.div>

      {/* Floating particles */}
      <FloatingParticles count={40} className="z-[3]" />

      {/* Network lines effect */}
      <NetworkLinesEffect className="z-[3] opacity-30" />

      {/* Parallax Decorative Elements */}
      <FloatingElement yRange={[-80, 80]} xRange={[-20, 20]} className="absolute top-20 left-10 w-24 h-24 opacity-40 hidden lg:block z-[15]">
        <FloatingShield className="w-full h-full" delay={0} />
      </FloatingElement>
      <FloatingElement yRange={[-60, 60]} rotateRange={[0, 180]} className="absolute top-40 right-20 w-32 h-32 opacity-25 hidden lg:block z-[15]">
        <OrbitRing className="w-full h-full" />
      </FloatingElement>
      <FloatingElement yRange={[-40, 40]} className="absolute bottom-40 left-20 w-32 h-64 opacity-20 hidden lg:block z-[15]">
        <DataStream className="w-full h-full" />
      </FloatingElement>
      <RotateOnScroll rotateRange={[0, 360]} className="absolute top-1/4 right-20 w-20 h-20 opacity-20 hidden xl:block z-[15]">
        <CryptoIcons className="w-full h-full" />
      </RotateOnScroll>

      {/* Morphing blobs - behind content */}
      <MorphingBlob className="absolute -top-20 -left-20 w-[500px] h-[500px] opacity-15 z-[4]" />
      <MorphingBlob className="absolute -bottom-40 -right-20 w-[400px] h-[400px] opacity-10 z-[4]" color="#60A5FA" />

      {/* Gradient Overlays */}
      <motion.div className="absolute inset-0 bg-gradient-to-b from-dark-950/80 via-dark-950/30 to-dark-950 z-[5]" style={{ opacity }} />

      {/* Content */}
      <div className="relative z-20 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-8 pb-20">                                   
        <div className="text-center max-w-5xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            {/* Badge */}
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2, duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 bg-gold-500/10 border border-gold-500/30 rounded-full mb-8"
            >
              <Sparkles className="w-4 h-4 text-gold-500" />
              <span className="text-gold-400 text-sm font-medium">Decentralized VPN Network</span>
            </motion.div>

            {/* Main Headline with staggered animation */}
            <h1 className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-black tracking-tight mb-8 leading-[0.9]">
              <ScrollReveal direction="up" delay={0.1}>
                <span className="text-gradient-white inline-block">The Future of</span>
              </ScrollReveal>
              <br />
              <ScrollReveal direction="up" delay={0.2}>
                <span className="text-gradient glow-text inline-block">Decentralized</span>
              </ScrollReveal>
              <br />
              <ScrollReveal direction="up" delay={0.3}>
                <span className="text-gradient-white inline-block">Privacy</span>
              </ScrollReveal>
            </h1>

            {/* Subheadline with typing effect */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 0.6 }}
              className="text-xl md:text-2xl text-gray-400 mb-12 max-w-3xl mx-auto leading-relaxed"
            >
              <p className="mb-2">Military-grade encryption meets a global peer-to-peer network.</p>
              <p>
                <span className="text-white">Browse anonymously</span> or{' '}
                <span className="text-gold-400">
                  <TypingText
                    texts={['earn crypto', 'earn ETH', 'earn BTC', 'run a node']}
                    className="text-gold-400"
                  />
                </span>
              </p>
            </motion.div>

            {/* CTA Buttons with magnetic effect */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.6, duration: 0.6 }}
              className="flex flex-col sm:flex-row justify-center gap-4 mb-16"
            >
              <MagneticElement strength={0.2}>
                <motion.a
                  href="/register"
                  className="group relative inline-flex items-center justify-center px-8 py-4 bg-gradient-to-r from-gold-500 to-gold-600 text-dark-950 text-lg font-bold rounded-full shadow-lg shadow-gold-500/25"
                  whileHover={{ scale: 1.05, boxShadow: "0 20px 40px rgba(245, 158, 11, 0.3)" }}
                  whileTap={{ scale: 0.98 }}
                >
                  <span className="flex items-center gap-2">
                    Start Browsing Free
                    <ArrowRight className="w-5 h-5 group-hover:translate-x-2 transition-transform" />
                  </span>
                </motion.a>
              </MagneticElement>
              <MagneticElement strength={0.2}>
                <motion.a
                  href="#earnings"
                  className="group inline-flex items-center justify-center px-8 py-4 glass rounded-full text-lg font-medium hover:bg-white/10 transition-all"
                  whileHover={{ scale: 1.02 }}
                >
                  <Coins className="w-5 h-5 mr-2 text-gold-500" />
                  Become an Operator
                </motion.a>
              </MagneticElement>
            </motion.div>

            {/* Trust Indicators */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.8, duration: 0.6 }}
              className="flex flex-wrap justify-center items-center gap-8 text-gray-500 text-sm"
            >
              <div className="flex items-center gap-2">
                <Lock className="w-4 h-4 text-green-500" />
                <span>Zero-Log Policy</span>
              </div>
              <div className="flex items-center gap-2">
                <Shield className="w-4 h-4 text-blue-500" />
                <span>WireGuard Protocol</span>
              </div>
              <div className="flex items-center gap-2">
                <Globe className="w-4 h-4 text-purple-500" />
                <span>60+ Countries</span>
              </div>
            </motion.div>
          </motion.div>
        </div>
      </div>

      {/* Scroll Indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.2, duration: 0.6 }}
        className="absolute bottom-8 left-1/2 -translate-x-1/2 z-20"
      >
        <a href="#stats" className="flex flex-col items-center gap-2 text-gray-500 hover:text-white transition-colors">
          <span className="text-xs uppercase tracking-wider">Scroll</span>
          <ChevronDown className="w-5 h-5 scroll-indicator" />
        </a>
      </motion.div>
    </section>
  );
}

// ============================================
// STATS SECTION
// ============================================

function StatsSection() {
  const [ref, inView] = useInViewObserver({ triggerOnce: true, threshold: 0.3 });

  const stats = [
    { value: 5000, suffix: '+', label: 'Active Nodes', icon: Server },
    { value: 60, suffix: '+', label: 'Countries', icon: MapPin },
    { value: 10, suffix: 'Gbps', label: 'Max Speed', icon: Zap },
    { value: 0.10, prefix: '$', suffix: '/GB', label: 'Operator Earnings', icon: Coins, decimals: 2 },
  ];

  return (
    <section id="stats" className="relative py-20 overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 bg-gradient-to-b from-dark-950 via-dark-900/50 to-dark-950" />
      <div className="absolute inset-0 bg-grid-pattern bg-grid opacity-5" />

      <div ref={ref} className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          variants={staggerContainer}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="grid grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8"
        >
          {stats.map((stat, index) => (
            <motion.div
              key={stat.label}
              variants={fadeInUp}
              className="relative group"
            >
              <div className="glass rounded-2xl p-6 lg:p-8 text-center hover:bg-white/10 transition-all duration-300">
                <div className="absolute inset-0 bg-gradient-to-br from-gold-500/5 to-transparent rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
                <stat.icon className="w-8 h-8 text-gold-500 mx-auto mb-4" />
                <div className="text-3xl lg:text-4xl font-black text-white mb-2 stat-number">
                  {inView && (
                    <CountUp
                      start={0}
                      end={stat.value}
                      duration={2.5}
                      decimals={stat.decimals || 0}
                      prefix={stat.prefix || ''}
                      suffix={stat.suffix || ''}
                    />
                  )}
                </div>
                <div className="text-gray-400 font-medium">{stat.label}</div>
              </div>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}

// ============================================
// FEATURES SECTION
// ============================================

function FeaturesSection() {
  const features = [
    {
      icon: Globe,
      title: "Decentralized Network",
      description: "No central points of failure. Our network is powered by thousands of independent node operators distributed worldwide.",
      color: "from-blue-500 to-cyan-500"
    },
    {
      icon: Coins,
      title: "Earn Cryptocurrency",
      description: "Turn your unused bandwidth into passive income. Run a node and get paid in ETH, BTC, or LTC automatically.",
      color: "from-gold-500 to-orange-500"
    },
    {
      icon: Lock,
      title: "Zero-Log Policy",
      description: "We couldn't track you if we wanted to. Our architecture ensures complete anonymity by design, not by promise.",
      color: "from-green-500 to-emerald-500"
    },
    {
      icon: Zap,
      title: "Lightning Fast",
      description: "WireGuard protocol delivers speeds up to 10Gbps. Experience no buffering or throttling on any connection.",
      color: "from-purple-500 to-pink-500"
    },
    {
      icon: Shield,
      title: "Military Encryption",
      description: "ChaCha20-Poly1305 encryption protects your data. Same standards used by government agencies worldwide.",
      color: "from-red-500 to-rose-500"
    },
    {
      icon: Network,
      title: "Multi-Hop Routing",
      description: "Route traffic through multiple nodes for extra anonymity. Perfect for journalists and activists in hostile regions.",
      color: "from-indigo-500 to-violet-500"
    }
  ];

  return (
    <section id="features" className="relative py-32 overflow-hidden">
      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 right-0 w-[400px] h-[400px] opacity-10">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=600&auto=format&fit=crop"
            alt="Cyber security"
            className="w-full h-full object-cover rounded-full blur-md"
            scale={1.3}
            orientation="down"
          />
        </div>
        <div className="absolute bottom-0 left-0 w-[350px] h-[350px] opacity-10">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1563986768609-322da13575f3?w=600&auto=format&fit=crop"
            alt="Lock security"
            className="w-full h-full object-cover rounded-full blur-md"
            scale={1.4}
            orientation="up"
          />
        </div>
      </div>

      {/* Background Elements */}
      <div className="absolute inset-0 mesh-gradient" />
      <GridBackground className="absolute inset-0 pointer-events-none z-0" />

      {/* Parallax Decorations */}
      <ParallaxLayer speed={0.4} className="absolute top-20 right-10 w-40 h-40 opacity-20 hidden lg:block">
        <ConnectionLines className="w-full h-full" />
      </ParallaxLayer>
      <ParallaxLayer speed={0.6} className="absolute bottom-20 left-10 w-24 h-24 opacity-25 hidden lg:block">
        <CryptoIcons className="w-full h-full" />
      </ParallaxLayer>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-20">
          <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
            Why Choose Aureo
          </span>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black mb-6">
            <span className="text-gradient-white">Privacy Without</span>
            <br />
            <span className="text-gradient">Compromise</span>
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Experience the next generation of VPN technology. Built for privacy advocates, by privacy advocates.
          </p>
        </AnimatedSection>

        <StaggeredContainer className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8" staggerDelay={0.15}>
          {features.map((feature, index) => (
            <StaggeredItem key={feature.title}>
              <TiltCard tiltStrength={8}>
                <div className="feature-card relative glass rounded-2xl p-8 h-full group cursor-pointer overflow-hidden border border-transparent hover:border-gold-500/20 transition-all duration-300">
                  {/* Gradient Background on Hover */}
                  <div className={`absolute inset-0 bg-gradient-to-br ${feature.color} opacity-0 group-hover:opacity-5 rounded-2xl transition-opacity duration-500`} />

                  {/* Icon with float effect */}
                  <motion.div
                    className={`relative w-14 h-14 rounded-xl bg-gradient-to-br ${feature.color} p-0.5 mb-6`}
                    whileHover={{ rotate: [0, -10, 10, 0], scale: 1.1 }}
                    transition={{ duration: 0.5 }}
                  >
                    <div className="w-full h-full bg-dark-900 rounded-xl flex items-center justify-center">
                      <feature.icon className="w-7 h-7 text-white" />
                    </div>
                  </motion.div>

                  {/* Content */}
                  <h3 className="text-xl font-bold text-white mb-3 group-hover:text-gold-400 transition-colors">
                    {feature.title}
                  </h3>
                  <p className="text-gray-400 leading-relaxed">
                    {feature.description}
                  </p>

                  {/* Animated Arrow */}
                  <motion.div
                    className="absolute bottom-8 right-8"
                    initial={{ opacity: 0, x: -10 }}
                    whileHover={{ opacity: 1, x: 0 }}
                  >
                    <ArrowUpRight className="w-5 h-5 text-gold-500" />
                  </motion.div>

                  {/* Shine effect on hover */}
                  <motion.div
                    className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-1000"
                  />
                </div>
              </TiltCard>
            </StaggeredItem>
          ))}
        </StaggeredContainer>
      </div>
    </section>
  );
}

// ============================================
// HOW IT WORKS SECTION
// ============================================

function HowItWorksSection() {
  const steps = [
    {
      number: "01",
      title: "Download & Install",
      description: "Get our lightweight app for Windows, macOS, Linux, iOS, or Android. Setup takes under 60 seconds.",
      icon: Cpu
    },
    {
      number: "02",
      title: "Connect Instantly",
      description: "Choose from 5,000+ nodes across 60 countries. One click and you're protected with military-grade encryption.",
      icon: Globe
    },
    {
      number: "03",
      title: "Browse Freely",
      description: "Access any content without restrictions. Your real IP is hidden, and your data is fully encrypted.",
      icon: Shield
    }
  ];

  return (
    <section id="how-it-works" className="relative py-32 overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-b from-dark-950 via-dark-900/30 to-dark-950" />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-20">
          <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
            Getting Started
          </span>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black mb-6">
            <span className="text-gradient-white">Start in</span>{' '}
            <span className="text-gradient">60 Seconds</span>
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            No complicated setup. No technical knowledge required. Just download and connect.
          </p>
        </AnimatedSection>

        {/* Steps Timeline */}
        <div className="relative">
          {/* Connection Line */}
          <div className="hidden lg:block absolute top-1/2 left-0 right-0 h-0.5 bg-gradient-to-r from-transparent via-gold-500/30 to-transparent transform -translate-y-1/2" />

          <div className="grid lg:grid-cols-3 gap-8 lg:gap-12">
            {steps.map((step, index) => (
              <AnimatedSection key={step.number} delay={index * 0.2}>
                <div className="relative text-center lg:text-left">
                  {/* Step Number */}
                  <div className="flex lg:justify-start justify-center mb-6">
                    <div className="relative">
                      <div className="absolute inset-0 bg-gold-500/20 blur-xl rounded-full" />
                      <div className="relative w-20 h-20 rounded-full bg-gradient-to-br from-gold-500 to-gold-600 flex items-center justify-center">
                        <span className="text-2xl font-black text-dark-950">{step.number}</span>
                      </div>
                    </div>
                  </div>

                  {/* Icon */}
                  <div className="glass w-14 h-14 rounded-xl flex items-center justify-center mb-4 mx-auto lg:mx-0">
                    <step.icon className="w-7 h-7 text-gold-500" />
                  </div>

                  {/* Content */}
                  <h3 className="text-2xl font-bold text-white mb-3">{step.title}</h3>
                  <p className="text-gray-400 leading-relaxed max-w-sm mx-auto lg:mx-0">{step.description}</p>
                </div>
              </AnimatedSection>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

// ============================================
// NETWORK VISUALIZATION SECTION
// ============================================

function NetworkSection() {
  return (
    <section className="relative py-32 overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 bg-dark-950" />

      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-20 left-1/4 w-[500px] h-[500px] opacity-8">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=800&auto=format&fit=crop"
            alt="Server infrastructure"
            className="w-full h-full object-cover rounded-full blur-lg"
            scale={1.5}
            orientation="up"
          />
        </div>
        <div className="absolute bottom-0 right-10 w-[400px] h-[400px] opacity-8 hidden lg:block">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1544197150-b99a580bb7a8?w=600&auto=format&fit=crop"
            alt="Network cables"
            className="w-full h-full object-cover rounded-2xl blur-md"
            scale={1.3}
            orientation="down"
          />
        </div>
      </div>

      {/* Parallax decorations */}
      <ParallaxLayer speed={0.5} className="absolute top-10 left-1/4 w-32 h-32 opacity-15 hidden lg:block">
        <OrbitRing className="w-full h-full" reverse />
      </ParallaxLayer>
      <ParallaxLayer speed={0.3} className="absolute bottom-20 right-10 w-20 h-40 opacity-20 hidden lg:block">
        <DataStream className="w-full h-full" />
      </ParallaxLayer>

      {/* Animated wave at bottom */}
      <div className="absolute bottom-0 left-0 right-0 h-32 pointer-events-none">
        <WavePattern className="w-full h-full" />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Left Content */}
          <AnimatedSection>
            <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
              Decentralized Architecture
            </span>
            <h2 className="text-4xl md:text-5xl font-black mb-6">
              <span className="text-gradient-white">A Network Built</span>
              <br />
              <span className="text-gradient">For Resilience</span>
            </h2>
            <p className="text-xl text-gray-400 mb-8 leading-relaxed">
              Unlike traditional VPNs with centralized servers, Aureo uses a peer-to-peer network
              that's impossible to shut down. Every node strengthens the network.
            </p>

            <div className="space-y-6">
              {[
                { icon: Network, title: "P2P Architecture", desc: "No single point of failure or control" },
                { icon: Shield, title: "End-to-End Encrypted", desc: "Your data is never visible to nodes" },
                { icon: Users, title: "Community Powered", desc: "5,000+ independent operators worldwide" }
              ].map((item, i) => (
                <div key={i} className="flex items-start gap-4">
                  <div className="glass rounded-lg p-3 flex-shrink-0">
                    <item.icon className="w-6 h-6 text-gold-500" />
                  </div>
                  <div>
                    <h4 className="font-bold text-white mb-1">{item.title}</h4>
                    <p className="text-gray-400">{item.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </AnimatedSection>

          {/* Right - 3D Network Visualization */}
          <ZoomOnScroll zoomRange={[0.8, 1]}>
            <div className="relative aspect-square max-w-lg mx-auto">
              <div className="absolute inset-0 bg-gradient-to-br from-gold-500/10 to-transparent rounded-3xl" />
              <div className="glass rounded-3xl overflow-hidden h-full relative">
                {/* Hexagon background */}
                <HexagonGrid className="opacity-30" />

                {/* 3D Network with data packets */}
                <div className="absolute inset-0">
                  <Suspense fallback={
                    <div className="w-full h-full flex items-center justify-center">
                      <div className="w-16 h-16 border-4 border-gold-500/30 border-t-gold-500 rounded-full animate-spin" />
                    </div>
                  }>
                    <NetworkVisualization3D />
                  </Suspense>
                </div>

                {/* Radar overlay */}
                <div className="absolute bottom-4 right-4 opacity-50">
                  <RadarSweep className="w-24 h-24" size={100} />
                </div>
              </div>
            </div>
          </ZoomOnScroll>
        </div>
      </div>
    </section>
  );
}

// ============================================
// EARNINGS SECTION
// ============================================

function EarningsSection() {
  const [bandwidth, setBandwidth] = useState(100);
  const [tier, setTier] = useState('gold');

  const tierMultipliers = {
    bronze: { multiplier: 1.0, label: 'Bronze', color: 'text-orange-400' },
    silver: { multiplier: 1.3, label: 'Silver', color: 'text-gray-300' },
    gold: { multiplier: 1.6, label: 'Gold', color: 'text-gold-400' },
    platinum: { multiplier: 2.0, label: 'Platinum', color: 'text-blue-300' }
  };

  const baseRate = 0.02; // $0.02 per GB
  const monthlyEarnings = bandwidth * 30 * baseRate * tierMultipliers[tier].multiplier;

  const chartData = useMemo(() => {
    return Array.from({ length: 12 }, (_, i) => ({
      month: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'][i],
      earnings: monthlyEarnings * (0.8 + Math.random() * 0.4)
    }));
  }, [monthlyEarnings]);

  return (
    <section id="earnings" className="relative py-32 overflow-hidden">
      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-10 right-0 w-[350px] h-[350px] opacity-10">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1621761191319-c6fb62004040?w=600&auto=format&fit=crop"
            alt="Cryptocurrency"
            className="w-full h-full object-cover rounded-full blur-md"
            scale={1.4}
            orientation="down"
          />
        </div>
        <div className="absolute bottom-10 left-0 w-[300px] h-[300px] opacity-10">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1516245834210-c4c142787335?w=600&auto=format&fit=crop"
            alt="Bitcoin"
            className="w-full h-full object-cover rounded-full blur-md"
            scale={1.3}
            orientation="up"
          />
        </div>
      </div>

      <div className="absolute inset-0 mesh-gradient" />
      <GridBackground className="absolute inset-0 pointer-events-none" />

      {/* Parallax crypto icons */}
      <ParallaxLayer speed={0.4} className="absolute top-20 left-10 w-32 h-32 opacity-30 hidden lg:block">
        <CryptoIcons className="w-full h-full" />
      </ParallaxLayer>
      <ParallaxLayer speed={0.6} className="absolute bottom-40 right-20 w-24 h-24 opacity-20 hidden lg:block">
        <FloatingShield className="w-full h-full" />
      </ParallaxLayer>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16">
          <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
            Node Operator Rewards
          </span>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black mb-6">
            <span className="text-gradient-white">Turn Bandwidth Into</span>
            <br />
            <span className="text-gradient">Passive Income</span>
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Run a node on your server or VPS. Earn cryptocurrency for every GB of traffic you route.
          </p>
        </AnimatedSection>

        <div className="grid lg:grid-cols-2 gap-12 items-start">
          {/* Calculator */}
          <AnimatedSection>
            <div className="glass rounded-2xl p-8">
              <h3 className="text-2xl font-bold text-white mb-6 flex items-center gap-3">
                <TrendingUp className="w-6 h-6 text-gold-500" />
                Earnings Calculator
              </h3>

              {/* Bandwidth Slider */}
              <div className="mb-8">
                <label className="block text-gray-400 mb-3">Daily Bandwidth (GB)</label>
                <input
                  type="range"
                  min="10"
                  max="500"
                  value={bandwidth}
                  onChange={(e) => setBandwidth(Number(e.target.value))}
                  className="w-full h-2 bg-dark-800 rounded-lg appearance-none cursor-pointer accent-gold-500"
                />
                <div className="flex justify-between mt-2 text-sm text-gray-500">
                  <span>10 GB</span>
                  <span className="text-gold-500 font-bold text-lg">{bandwidth} GB/day</span>
                  <span>500 GB</span>
                </div>
              </div>

              {/* Tier Selection */}
              <div className="mb-8">
                <label className="block text-gray-400 mb-3">Your Tier</label>
                <div className="grid grid-cols-4 gap-2">
                  {Object.entries(tierMultipliers).map(([key, val]) => (
                    <button
                      key={key}
                      onClick={() => setTier(key)}
                      className={`py-3 px-4 rounded-lg font-medium transition-all ${
                        tier === key
                          ? 'bg-gold-500 text-dark-950'
                          : 'glass hover:bg-white/10'
                      }`}
                    >
                      <span className={tier === key ? '' : val.color}>{val.label}</span>
                    </button>
                  ))}
                </div>
              </div>

              {/* Estimated Earnings */}
              <div className="bg-gradient-to-br from-gold-500/20 to-gold-600/10 rounded-xl p-6 border border-gold-500/20">
                <div className="text-gray-400 mb-2">Estimated Monthly Earnings</div>
                <div className="text-5xl font-black text-gold-400 mb-2">
                  ${monthlyEarnings.toFixed(2)}
                </div>
                <div className="text-gray-500 text-sm">
                  Based on {bandwidth * 30} GB/month at {tierMultipliers[tier].multiplier}x multiplier
                </div>
              </div>
            </div>
          </AnimatedSection>

          {/* Chart */}
          <AnimatedSection delay={0.2}>
            <div className="glass rounded-2xl p-8">
              <h3 className="text-2xl font-bold text-white mb-6 flex items-center gap-3">
                <Activity className="w-6 h-6 text-gold-500" />
                Projected Annual Earnings
              </h3>

              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="colorEarnings" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#F59E0B" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#F59E0B" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <XAxis dataKey="month" stroke="#6b7280" fontSize={12} />
                    <YAxis stroke="#6b7280" fontSize={12} tickFormatter={(val) => `$${val.toFixed(0)}`} />
                    <Tooltip
                      contentStyle={{ background: '#1f2937', border: 'none', borderRadius: '8px' }}
                      labelStyle={{ color: '#9ca3af' }}
                      formatter={(value) => [`$${value.toFixed(2)}`, 'Earnings']}
                    />
                    <Area
                      type="monotone"
                      dataKey="earnings"
                      stroke="#F59E0B"
                      strokeWidth={2}
                      fillOpacity={1}
                      fill="url(#colorEarnings)"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>

              {/* Payment Methods */}
              <div className="mt-6 pt-6 border-t border-white/10">
                <div className="text-gray-400 text-sm mb-3">Payout Options</div>
                <div className="flex gap-4">
                  {[
                    { icon: CircleDollarSign, label: 'ETH' },
                    { icon: Bitcoin, label: 'BTC' },
                    { icon: Coins, label: 'LTC' }
                  ].map((coin) => (
                    <div key={coin.label} className="flex items-center gap-2 glass rounded-lg px-4 py-2">
                      <coin.icon className="w-5 h-5 text-gold-500" />
                      <span className="text-white font-medium">{coin.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </AnimatedSection>
        </div>
      </div>
    </section>
  );
}

// ============================================
// PRICING SECTION
// ============================================

function PricingSection() {
  const [annual, setAnnual] = useState(true);

  const plans = [
    {
      name: 'Free',
      description: 'Get started with basic protection',
      price: { monthly: 0, annual: 0 },
      features: [
        '1 Device',
        '5 Server Locations',
        '500 MB/day Limit',
        'Basic Support',
        'Standard Speed'
      ],
      notIncluded: ['Kill Switch', 'Multi-Hop', 'Dedicated IP'],
      cta: 'Start Free',
      popular: false
    },
    {
      name: 'Pro',
      description: 'For serious privacy enthusiasts',
      price: { monthly: 9.99, annual: 4.99 },
      features: [
        '5 Devices',
        'All 60+ Locations',
        'Unlimited Bandwidth',
        'Priority Support',
        'Maximum Speed',
        'Kill Switch',
        'Split Tunneling'
      ],
      notIncluded: ['Dedicated IP'],
      cta: 'Get Pro',
      popular: true
    },
    {
      name: 'Business',
      description: 'For teams and enterprises',
      price: { monthly: 14.99, annual: 9.99 },
      features: [
        'Unlimited Devices',
        'All 60+ Locations',
        'Unlimited Bandwidth',
        '24/7 Support',
        'Maximum Speed',
        'Kill Switch',
        'Multi-Hop Routing',
        'Dedicated IP',
        'Admin Dashboard'
      ],
      notIncluded: [],
      cta: 'Contact Sales',
      popular: false
    }
  ];

  return (
    <section id="pricing" className="relative py-32 overflow-hidden">
      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-20 left-0 w-[300px] h-[300px] opacity-8">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=600&auto=format&fit=crop"
            alt="Blockchain"
            className="w-full h-full object-cover rounded-full blur-lg"
            scale={1.3}
            orientation="up"
          />
        </div>
        <div className="absolute bottom-20 right-0 w-[280px] h-[280px] opacity-8">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1620321023374-d1a68fbc720d?w=600&auto=format&fit=crop"
            alt="Ethereum"
            className="w-full h-full object-cover rounded-full blur-lg"
            scale={1.4}
            orientation="down"
          />
        </div>
      </div>

      <div className="absolute inset-0 bg-gradient-to-b from-dark-950 via-dark-900/30 to-dark-950" />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16">
          <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
            Simple Pricing
          </span>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black mb-6">
            <span className="text-gradient-white">Choose Your</span>{' '}
            <span className="text-gradient">Plan</span>
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto mb-8">
            Start free forever. Upgrade when you need more features.
          </p>

          {/* Billing Toggle */}
          <div className="inline-flex items-center gap-4 glass rounded-full p-1.5">
            <button
              onClick={() => setAnnual(false)}
              className={`px-6 py-2 rounded-full font-medium transition-all ${
                !annual ? 'bg-gold-500 text-dark-950' : 'text-gray-400 hover:text-white'
              }`}
            >
              Monthly
            </button>
            <button
              onClick={() => setAnnual(true)}
              className={`px-6 py-2 rounded-full font-medium transition-all ${
                annual ? 'bg-gold-500 text-dark-950' : 'text-gray-400 hover:text-white'
              }`}
            >
              Annual
              <span className="ml-2 text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">
                Save 50%
              </span>
            </button>
          </div>
        </AnimatedSection>

        <StaggeredContainer className="grid md:grid-cols-3 gap-8" staggerDelay={0.15}>
          {plans.map((plan, index) => (
            <StaggeredItem key={plan.name}>
              <TiltCard tiltStrength={5}>
                <div className={`relative glass rounded-2xl p-8 h-full ${
                  plan.popular ? 'ring-2 ring-gold-500 shadow-lg shadow-gold-500/10' : ''
                }`}>

                  {plan.popular && (
                    <motion.div
                      className="absolute -top-4 left-1/2 -translate-x-1/2"
                      animate={{ y: [0, -5, 0] }}
                      transition={{ duration: 2, repeat: Infinity }}
                    >
                      <span className="bg-gradient-to-r from-gold-500 to-gold-600 text-dark-950 text-sm font-bold px-4 py-1.5 rounded-full shadow-lg shadow-gold-500/30">
                        Most Popular
                      </span>
                    </motion.div>
                  )}

                  <div className="text-center mb-8 relative z-10">
                    <h3 className="text-2xl font-bold text-white mb-2">{plan.name}</h3>
                    <p className="text-gray-400 mb-4">{plan.description}</p>
                    <div className="flex items-baseline justify-center gap-1">
                      <motion.span
                        className="text-5xl font-black text-white"
                        key={annual ? 'annual' : 'monthly'}
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.3 }}
                      >
                        ${annual ? plan.price.annual : plan.price.monthly}
                      </motion.span>
                      <span className="text-gray-400">/mo</span>
                    </div>
                    {annual && plan.price.annual > 0 && (
                      <motion.div
                        className="text-gray-500 text-sm mt-1"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.2 }}
                      >
                        Billed annually (${(plan.price.annual * 12).toFixed(0)}/year)
                      </motion.div>
                    )}
                  </div>

                  <ul className="space-y-3 mb-8 relative z-10">
                    {plan.features.map((feature, i) => (
                      <motion.li
                        key={feature}
                        className="flex items-center gap-3 text-gray-300"
                        initial={{ opacity: 0, x: -20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.05 }}
                        viewport={{ once: true }}
                      >
                        <motion.div
                          whileHover={{ scale: 1.2, rotate: 360 }}
                          transition={{ duration: 0.3 }}
                        >
                          <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
                        </motion.div>
                        {feature}
                      </motion.li>
                    ))}
                    {plan.notIncluded.map((feature) => (
                      <li key={feature} className="flex items-center gap-3 text-gray-500">
                        <X className="w-5 h-5 text-gray-600 flex-shrink-0" />
                        {feature}
                      </li>
                    ))}
                  </ul>

                  <MagneticElement strength={0.15}>
                    <motion.a
                      href="/register"
                      className={`relative block w-full py-3 rounded-xl font-bold text-center transition-all overflow-hidden ${
                        plan.popular
                          ? 'bg-gradient-to-r from-gold-500 to-gold-600 text-dark-950 hover:opacity-90'
                          : 'glass hover:bg-white/10'
                      }`}
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                    >
                      <span className="relative z-10">{plan.cta}</span>
                      <motion.div
                        className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent -translate-x-full"
                        whileHover={{ translateX: '200%' }}
                        transition={{ duration: 0.6 }}
                      />
                    </motion.a>
                  </MagneticElement>

                  {/* Background glow for popular */}
                  {plan.popular && (
                    <div className="absolute inset-0 bg-gradient-to-br from-gold-500/5 to-transparent pointer-events-none" />
                  )}
                </div>
              </TiltCard>
            </StaggeredItem>
          ))}
        </StaggeredContainer>
      </div>
    </section>
  );
}

// ============================================
// TESTIMONIALS SECTION
// ============================================

function TestimonialsSection() {
  const sectionRef = useRef(null);
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start end', 'end start']
  });
  const parallaxY = useTransform(scrollYProgress, [0, 1], [50, -50]);

  const testimonials = [
    {
      quote: "Finally a VPN that doesn't log my data. The decentralized approach gives me peace of mind knowing there's no central server to subpoena.",
      author: "Sarah Chen",
      role: "Security Researcher",
      avatar: "SC"
    },
    {
      quote: "I've made over $500/month running 3 nodes. The setup was incredibly easy, and payouts arrive like clockwork every week.",
      author: "Marcus Rodriguez",
      role: "Node Operator",
      avatar: "MR"
    },
    {
      quote: "The speeds are insane. I was skeptical about a decentralized VPN being fast, but WireGuard really delivers. No buffering on 4K streams.",
      author: "James Wilson",
      role: "Digital Nomad",
      avatar: "JW"
    }
  ];

  return (
    <section ref={sectionRef} className="relative py-32 overflow-hidden">
      {/* Parallax Background Images */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-0 w-[300px] h-[300px] opacity-8">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1522071820081-009f0129c71c?w=600&auto=format&fit=crop"
            alt="Team collaboration"
            className="w-full h-full object-cover rounded-full blur-lg"
            scale={1.3}
            orientation="up"
          />
        </div>
        <div className="absolute bottom-1/4 right-0 w-[250px] h-[250px] opacity-8">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1553877522-43269d4ea984?w=600&auto=format&fit=crop"
            alt="Digital workspace"
            className="w-full h-full object-cover rounded-full blur-lg"
            scale={1.4}
            orientation="down"
          />
        </div>
      </div>

      <div className="absolute inset-0 mesh-gradient" />

      {/* Parallax background elements */}
      <motion.div style={{ y: parallaxY }} className="absolute inset-0 pointer-events-none">
        <div className="absolute top-20 left-10 w-24 h-24 opacity-20 hidden lg:block">
          <OrbitRing className="w-full h-full" />
        </div>
        <div className="absolute bottom-20 right-10 w-32 h-32 opacity-15 hidden lg:block">
          <FloatingShield className="w-full h-full" delay={0.5} />
        </div>
      </motion.div>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16">
          <span className="inline-block text-gold-500 font-semibold tracking-wider uppercase text-sm mb-4">
            Testimonials
          </span>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black mb-6">
            <span className="text-gradient-white">Loved by</span>{' '}
            <span className="text-gradient">Privacy Advocates</span>
          </h2>
        </AnimatedSection>

        <div className="grid md:grid-cols-3 gap-8">
          {testimonials.map((testimonial, index) => (
            <AnimatedSection key={testimonial.author} delay={index * 0.1}>
              <MouseParallax strength={10}>
                <div className="glass rounded-2xl p-8 h-full flex flex-col hover:bg-white/10 transition-all duration-300">
                  <Quote className="w-10 h-10 text-gold-500/30 mb-4" />
                  <p className="text-gray-300 text-lg leading-relaxed mb-6 flex-grow">
                    "{testimonial.quote}"
                  </p>
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gold-500 to-gold-600 flex items-center justify-center font-bold text-dark-950">
                      {testimonial.avatar}
                    </div>
                    <div>
                      <div className="font-bold text-white">{testimonial.author}</div>
                      <div className="text-gray-400 text-sm">{testimonial.role}</div>
                    </div>
                  </div>
                  <div className="flex gap-1 mt-4">
                    {[...Array(5)].map((_, i) => (
                      <Star key={i} className="w-4 h-4 fill-gold-500 text-gold-500" />
                    ))}
                  </div>
                </div>
              </MouseParallax>
            </AnimatedSection>
          ))}
        </div>
      </div>
    </section>
  );
}

// ============================================
// CTA SECTION
// ============================================

function CTASection() {
  const sectionRef = useRef(null);
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start end', 'end start']
  });
  const scale = useTransform(scrollYProgress, [0, 0.5], [0.9, 1]);
  const opacity = useTransform(scrollYProgress, [0, 0.3], [0, 1]);

  return (
    <motion.section
      ref={sectionRef}
      style={{ scale, opacity }}
      className="relative py-32 overflow-hidden"
    >
      {/* Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-gold-600 via-gold-500 to-gold-600" />
      <div className="absolute inset-0 bg-grid-pattern bg-grid opacity-10" />

      {/* Parallax Background Image */}
      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-20 -left-20 w-[400px] h-[400px] opacity-20">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=600&auto=format&fit=crop"
            alt="Matrix code"
            className="w-full h-full object-cover rounded-full mix-blend-overlay"
            scale={1.5}
            orientation="down"
          />
        </div>
        <div className="absolute -bottom-20 -right-20 w-[350px] h-[350px] opacity-20">
          <ParallaxImage
            src="https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=600&auto=format&fit=crop"
            alt="Cyber pattern"
            className="w-full h-full object-cover rounded-full mix-blend-overlay"
            scale={1.4}
            orientation="up"
          />
        </div>
      </div>

      {/* Animated Orbs with parallax */}
      <motion.div
        className="absolute top-0 left-1/4 w-96 h-96 bg-white/10 rounded-full blur-3xl"
        animate={{
          y: [0, -30, 0],
          x: [0, 20, 0],
        }}
        transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
      />
      <motion.div
        className="absolute bottom-0 right-1/4 w-72 h-72 bg-white/10 rounded-full blur-3xl"
        animate={{
          y: [0, 30, 0],
          x: [0, -20, 0],
        }}
        transition={{ duration: 10, repeat: Infinity, ease: "easeInOut" }}
      />

      <div className="relative z-10 max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
        <AnimatedSection>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-black text-dark-950 mb-6">
            Ready to Reclaim Your Privacy?
          </h2>
          <p className="text-xl text-dark-800 mb-10 max-w-2xl mx-auto">
            Join thousands of users who trust Aureo for secure, decentralized internet access.
            Start free, upgrade when you're ready.
          </p>
          <div className="flex flex-col sm:flex-row justify-center gap-4">
            <motion.a
              href="/register"
              className="inline-flex items-center justify-center px-8 py-4 bg-dark-950 text-white text-lg font-bold rounded-full shadow-xl"
              whileHover={{ scale: 1.05, boxShadow: "0 20px 40px rgba(0,0,0,0.3)" }}
              whileTap={{ scale: 0.98 }}
            >
              Get Started Free
              <ArrowRight className="ml-2 w-5 h-5" />
            </motion.a>
            <motion.a
              href="#earnings"
              className="inline-flex items-center justify-center px-8 py-4 bg-white/20 backdrop-blur-sm text-dark-950 text-lg font-bold rounded-full border border-white/30"
              whileHover={{ scale: 1.02, background: "rgba(255,255,255,0.3)" }}
            >
              <Coins className="mr-2 w-5 h-5" />
              Become an Operator
            </motion.a>
          </div>
        </AnimatedSection>
      </div>
    </motion.section>
  );
}

// ============================================
// FOOTER
// ============================================

function Footer() {
  const footerLinks = {
    Product: [
      { label: 'Features', path: '/features' },
      { label: 'Pricing', path: '/#pricing' },
      { label: 'Node Operators', path: '/node-operators' },
      { label: 'Downloads', path: '/downloads' },
    ],
    Resources: [
      { label: 'Documentation', path: '/docs' },
      { label: 'API Reference', path: '/api' },
      { label: 'Status', path: '/status' },
    ],
    Company: [
      { label: 'About', path: '/about' },
      { label: 'Contact', path: '/contact' },
    ],
    Legal: [
      { label: 'Privacy Policy', path: '/privacy' },
      { label: 'Terms of Service', path: '/terms' },
    ]
  };

  return (
    <footer className="relative bg-dark-950 border-t border-white/5 pt-20 pb-10 overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 bg-grid-pattern bg-grid opacity-5" />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Main Footer Content */}
        <div className="grid grid-cols-2 md:grid-cols-6 gap-8 mb-16">
          {/* Logo & Description */}
          <div className="col-span-2">
            <Link to="/" className="flex items-center gap-3 mb-4">
              <Shield className="h-8 w-8 text-gold-500" />
              <span className="font-bold text-xl">
                <span className="text-white">Aureo</span>
                <span className="text-gold-500">VPN</span>
              </span>
            </Link>
            <p className="text-gray-400 mb-6 max-w-sm">
              The future of decentralized privacy. Secure browsing and crypto rewards for the privacy-conscious.
            </p>
            <div className="flex gap-4">
              {[Twitter, Github, MessageCircle, Mail].map((Icon, i) => (
                <a
                  key={i}
                  href="#"
                  className="w-10 h-10 glass rounded-lg flex items-center justify-center hover:bg-white/10 transition-all"
                >
                  <Icon className="w-5 h-5 text-gray-400 hover:text-white" />
                </a>
              ))}
            </div>
          </div>

          {/* Links */}
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h4 className="font-bold text-white mb-4">{category}</h4>
              <ul className="space-y-3">
                {links.map((link) => (
                  <li key={link.path}>
                    <Link to={link.path} className="text-gray-400 hover:text-white transition-colors text-sm">
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-white/5 pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
          <div className="text-gray-500 text-sm">
            &copy; {new Date().getFullYear()} Aureo Network. All rights reserved.
          </div>
          <div className="flex items-center gap-6 text-sm text-gray-500">
            <Link to="/status" className="flex items-center gap-2 hover:text-white transition-colors">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              All systems operational
            </Link>
            <span>v2.0.0</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

export default App;
