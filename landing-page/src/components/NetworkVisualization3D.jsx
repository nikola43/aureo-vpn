import { useRef, useMemo, useState, useEffect } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { Line, Sphere, Trail, Float, Stars, useTexture } from '@react-three/drei';
import * as THREE from 'three';

// Animated data packet traveling along a path
function DataPacket({ startPoint, endPoint, speed = 1, delay = 0, color = "#F59E0B" }) {
  const ref = useRef();
  const [progress, setProgress] = useState(0);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setVisible(true), delay * 1000);
    return () => clearTimeout(timer);
  }, [delay]);

  useFrame((state, delta) => {
    if (!visible) return;
    setProgress((prev) => {
      const next = prev + delta * speed * 0.3;
      return next > 1 ? 0 : next;
    });
  });

  const position = useMemo(() => {
    const start = new THREE.Vector3(...startPoint);
    const end = new THREE.Vector3(...endPoint);
    return start.lerp(end, progress);
  }, [startPoint, endPoint, progress]);

  if (!visible) return null;

  return (
    <group position={[position.x, position.y, position.z]}>
      {/* Glow effect */}
      <mesh>
        <sphereGeometry args={[0.08, 16, 16]} />
        <meshBasicMaterial color={color} transparent opacity={0.3} />
      </mesh>
      {/* Core */}
      <mesh>
        <sphereGeometry args={[0.04, 16, 16]} />
        <meshBasicMaterial color={color} />
      </mesh>
      {/* Outer glow */}
      <pointLight color={color} intensity={2} distance={0.5} />
    </group>
  );
}

// Glowing connection line between nodes
function GlowingLine({ start, end, color = "#F59E0B", opacity = 0.4 }) {
  const ref = useRef();

  useFrame((state) => {
    if (ref.current) {
      ref.current.material.opacity = 0.2 + Math.sin(state.clock.elapsedTime * 2) * 0.2;
    }
  });

  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  return (
    <Line
      ref={ref}
      points={points}
      color={color}
      lineWidth={1.5}
      transparent
      opacity={opacity}
      dashed
      dashScale={5}
      dashSize={0.1}
      dashOffset={0}
    />
  );
}

// Network node with pulsing effect
function NetworkNode({ position, size = 0.1, color = "#F59E0B", pulseSpeed = 1 }) {
  const meshRef = useRef();
  const glowRef = useRef();

  useFrame((state) => {
    const pulse = Math.sin(state.clock.elapsedTime * pulseSpeed) * 0.3 + 1;
    if (meshRef.current) {
      meshRef.current.scale.setScalar(pulse);
    }
    if (glowRef.current) {
      glowRef.current.scale.setScalar(pulse * 1.5);
      glowRef.current.material.opacity = 0.2 + Math.sin(state.clock.elapsedTime * pulseSpeed) * 0.1;
    }
  });

  return (
    <Float speed={2} rotationIntensity={0.5} floatIntensity={0.5}>
      <group position={position}>
        {/* Outer glow */}
        <mesh ref={glowRef}>
          <sphereGeometry args={[size * 2, 16, 16]} />
          <meshBasicMaterial color={color} transparent opacity={0.2} />
        </mesh>
        {/* Core node */}
        <mesh ref={meshRef}>
          <sphereGeometry args={[size, 32, 32]} />
          <meshStandardMaterial
            color={color}
            emissive={color}
            emissiveIntensity={0.5}
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
        <pointLight color={color} intensity={1} distance={1} />
      </group>
    </Float>
  );
}

// Central hub node
function CentralHub({ position = [0, 0, 0] }) {
  const meshRef = useRef();
  const ringRef = useRef();

  useFrame((state) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += 0.005;
    }
    if (ringRef.current) {
      ringRef.current.rotation.z += 0.01;
      ringRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.5) * 0.2;
    }
  });

  return (
    <group position={position}>
      {/* Rotating ring */}
      <mesh ref={ringRef}>
        <torusGeometry args={[0.4, 0.02, 16, 64]} />
        <meshBasicMaterial color="#F59E0B" />
      </mesh>

      {/* Second ring */}
      <mesh rotation={[Math.PI / 2, 0, 0]}>
        <torusGeometry args={[0.35, 0.015, 16, 64]} />
        <meshBasicMaterial color="#F59E0B" transparent opacity={0.5} />
      </mesh>

      {/* Core sphere */}
      <mesh ref={meshRef}>
        <icosahedronGeometry args={[0.25, 1]} />
        <meshStandardMaterial
          color="#1f2937"
          emissive="#F59E0B"
          emissiveIntensity={0.3}
          metalness={0.9}
          roughness={0.1}
          wireframe
        />
      </mesh>

      {/* Inner glow */}
      <mesh>
        <sphereGeometry args={[0.2, 32, 32]} />
        <meshBasicMaterial color="#F59E0B" transparent opacity={0.3} />
      </mesh>

      <pointLight color="#F59E0B" intensity={3} distance={3} />
    </group>
  );
}

// Full network scene
function NetworkScene() {
  // Define network nodes positions
  const nodes = useMemo(() => [
    { position: [2, 1, 0], delay: 0 },
    { position: [-2, 1, 0.5], delay: 0.3 },
    { position: [1.5, -1, 1], delay: 0.6 },
    { position: [-1.5, -1, -0.5], delay: 0.9 },
    { position: [0, 2, -1], delay: 1.2 },
    { position: [0, -2, 0.5], delay: 1.5 },
    { position: [2.5, 0, -1], delay: 0.2 },
    { position: [-2.5, 0, 1], delay: 0.5 },
  ], []);

  const center = [0, 0, 0];

  return (
    <>
      {/* Ambient lighting */}
      <ambientLight intensity={0.3} />
      <pointLight position={[10, 10, 10]} intensity={0.5} />

      {/* Stars background */}
      <Stars radius={100} depth={50} count={2000} factor={4} fade speed={1} />

      {/* Central hub */}
      <CentralHub position={center} />

      {/* Network nodes */}
      {nodes.map((node, i) => (
        <NetworkNode
          key={i}
          position={node.position}
          size={0.08}
          pulseSpeed={1 + Math.random()}
        />
      ))}

      {/* Connection lines */}
      {nodes.map((node, i) => (
        <GlowingLine
          key={`line-${i}`}
          start={center}
          end={node.position}
        />
      ))}

      {/* Data packets traveling along connections */}
      {nodes.map((node, i) => (
        <DataPacket
          key={`packet-${i}`}
          startPoint={center}
          endPoint={node.position}
          speed={0.8 + Math.random() * 0.4}
          delay={node.delay}
        />
      ))}

      {/* Reverse packets (responses) */}
      {nodes.slice(0, 4).map((node, i) => (
        <DataPacket
          key={`packet-reverse-${i}`}
          startPoint={node.position}
          endPoint={center}
          speed={0.6 + Math.random() * 0.3}
          delay={node.delay + 1.5}
          color="#60A5FA"
        />
      ))}
    </>
  );
}

// Main component
export default function NetworkVisualization3D({ className = '' }) {
  return (
    <div className={`w-full h-full ${className}`}>
      <Canvas
        camera={{ position: [0, 0, 5], fov: 50 }}
        dpr={[1, 2]}
      >
        <NetworkScene />
      </Canvas>
    </div>
  );
}
