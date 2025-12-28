import { useRef, useEffect, useState } from 'react';
import { motion, useScroll, useTransform, useSpring, useInView } from 'framer-motion';

// Reveal animation that slides in from any direction
export function ScrollReveal({
  children,
  className = '',
  direction = 'up', // up, down, left, right
  delay = 0,
  duration = 0.8,
  distance = 100,
  once = true
}) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once, margin: "-100px" });

  const getInitialPosition = () => {
    switch (direction) {
      case 'up': return { y: distance, x: 0 };
      case 'down': return { y: -distance, x: 0 };
      case 'left': return { x: distance, y: 0 };
      case 'right': return { x: -distance, y: 0 };
      default: return { y: distance, x: 0 };
    }
  };

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, ...getInitialPosition() }}
      animate={isInView ? { opacity: 1, x: 0, y: 0 } : { opacity: 0, ...getInitialPosition() }}
      transition={{ duration, delay, ease: [0.25, 0.46, 0.45, 0.94] }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Staggered children animation
export function StaggeredContainer({
  children,
  className = '',
  staggerDelay = 0.1,
  initialDelay = 0.2
}) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-50px" });

  return (
    <motion.div
      ref={ref}
      className={className}
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
      variants={{
        hidden: { opacity: 0 },
        visible: {
          opacity: 1,
          transition: {
            staggerChildren: staggerDelay,
            delayChildren: initialDelay
          }
        }
      }}
    >
      {children}
    </motion.div>
  );
}

export function StaggeredItem({ children, className = '' }) {
  return (
    <motion.div
      className={className}
      variants={{
        hidden: { opacity: 0, y: 50, scale: 0.9 },
        visible: {
          opacity: 1,
          y: 0,
          scale: 1,
          transition: { duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }
        }
      }}
    >
      {children}
    </motion.div>
  );
}

// Parallax container with multiple layers
export function ParallaxContainer({ children, className = '' }) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start']
  });

  return (
    <div ref={ref} className={`relative ${className}`}>
      {children}
    </div>
  );
}

// Floating element that moves with scroll
export function FloatingElement({
  children,
  className = '',
  yRange = [-50, 50],
  xRange = [0, 0],
  rotateRange = [0, 0],
  scaleRange = [1, 1]
}) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start']
  });

  const y = useTransform(scrollYProgress, [0, 1], yRange);
  const x = useTransform(scrollYProgress, [0, 1], xRange);
  const rotate = useTransform(scrollYProgress, [0, 1], rotateRange);
  const scale = useTransform(scrollYProgress, [0, 1], scaleRange);

  const springY = useSpring(y, { stiffness: 100, damping: 30 });
  const springX = useSpring(x, { stiffness: 100, damping: 30 });

  return (
    <motion.div
      ref={ref}
      style={{ y: springY, x: springX, rotate, scale }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Text that reveals character by character
export function AnimatedText({ text, className = '', delay = 0 }) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true });
  const words = text.split(' ');

  return (
    <motion.span
      ref={ref}
      className={className}
    >
      {words.map((word, wordIndex) => (
        <span key={wordIndex} className="inline-block mr-2">
          {word.split('').map((char, charIndex) => (
            <motion.span
              key={charIndex}
              initial={{ opacity: 0, y: 50 }}
              animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 50 }}
              transition={{
                duration: 0.5,
                delay: delay + (wordIndex * 0.1) + (charIndex * 0.03),
                ease: [0.25, 0.46, 0.45, 0.94]
              }}
              className="inline-block"
            >
              {char}
            </motion.span>
          ))}
        </span>
      ))}
    </motion.span>
  );
}

// Counter that counts up when in view
export function ScrollCounter({
  end,
  duration = 2,
  className = '',
  prefix = '',
  suffix = '',
  decimals = 0
}) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true });
  const [count, setCount] = useState(0);

  useEffect(() => {
    if (!isInView) return;

    let startTime;
    const animate = (timestamp) => {
      if (!startTime) startTime = timestamp;
      const progress = Math.min((timestamp - startTime) / (duration * 1000), 1);
      setCount(progress * end);

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [isInView, end, duration]);

  return (
    <span ref={ref} className={className}>
      {prefix}{count.toFixed(decimals)}{suffix}
    </span>
  );
}

// Magnetic hover effect
export function MagneticElement({ children, className = '', strength = 0.3 }) {
  const ref = useRef(null);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const handleMouseMove = (e) => {
    if (!ref.current) return;
    const rect = ref.current.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    const x = (e.clientX - centerX) * strength;
    const y = (e.clientY - centerY) * strength;
    setPosition({ x, y });
  };

  const handleMouseLeave = () => {
    setPosition({ x: 0, y: 0 });
  };

  return (
    <motion.div
      ref={ref}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      animate={{ x: position.x, y: position.y }}
      transition={{ type: "spring", stiffness: 150, damping: 15 }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Tilt card effect
export function TiltCard({ children, className = '', tiltStrength = 10 }) {
  const ref = useRef(null);
  const [rotateX, setRotateX] = useState(0);
  const [rotateY, setRotateY] = useState(0);

  const handleMouseMove = (e) => {
    if (!ref.current) return;
    const rect = ref.current.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    const rotateX = ((e.clientY - centerY) / (rect.height / 2)) * -tiltStrength;
    const rotateY = ((e.clientX - centerX) / (rect.width / 2)) * tiltStrength;
    setRotateX(rotateX);
    setRotateY(rotateY);
  };

  const handleMouseLeave = () => {
    setRotateX(0);
    setRotateY(0);
  };

  return (
    <motion.div
      ref={ref}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      animate={{ rotateX, rotateY }}
      transition={{ type: "spring", stiffness: 300, damping: 30 }}
      style={{ transformStyle: 'preserve-3d', perspective: 1000 }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Zoom on scroll
export function ZoomOnScroll({ children, className = '', zoomRange = [0.8, 1] }) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'center center']
  });

  const scale = useTransform(scrollYProgress, [0, 1], zoomRange);
  const opacity = useTransform(scrollYProgress, [0, 0.5], [0, 1]);

  return (
    <motion.div
      ref={ref}
      style={{ scale, opacity }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Rotate on scroll
export function RotateOnScroll({
  children,
  className = '',
  rotateRange = [0, 360],
  axis = 'z'
}) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start']
  });

  const rotation = useTransform(scrollYProgress, [0, 1], rotateRange);

  const getRotationStyle = () => {
    switch (axis) {
      case 'x': return { rotateX: rotation };
      case 'y': return { rotateY: rotation };
      default: return { rotate: rotation };
    }
  };

  return (
    <motion.div
      ref={ref}
      style={getRotationStyle()}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Blur on scroll
export function BlurOnScroll({ children, className = '' }) {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'center center']
  });

  const blur = useTransform(scrollYProgress, [0, 1], [10, 0]);
  const opacity = useTransform(scrollYProgress, [0, 1], [0.5, 1]);

  return (
    <motion.div
      ref={ref}
      style={{
        filter: blur.get() > 0 ? `blur(${blur.get()}px)` : 'none',
        opacity
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Split text animation
export function SplitTextReveal({ text, className = '' }) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true });
  const lines = text.split('\n');

  return (
    <div ref={ref} className={className}>
      {lines.map((line, lineIndex) => (
        <div key={lineIndex} className="overflow-hidden">
          <motion.div
            initial={{ y: '100%' }}
            animate={isInView ? { y: 0 } : { y: '100%' }}
            transition={{
              duration: 0.8,
              delay: lineIndex * 0.15,
              ease: [0.25, 0.46, 0.45, 0.94]
            }}
          >
            {line}
          </motion.div>
        </div>
      ))}
    </div>
  );
}
