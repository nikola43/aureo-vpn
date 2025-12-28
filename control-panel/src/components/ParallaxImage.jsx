import { useEffect, useRef } from 'react';
import SimpleParallax from 'simple-parallax-js';

export default function ParallaxImage({
  src,
  alt = '',
  className = '',
  scale = 1.3,
  orientation = 'up',
  overflow = false,
  delay = 0.4,
  transition = 'cubic-bezier(0,0,0,1)',
  maxTransition = 60
}) {
  const imageRef = useRef(null);
  const parallaxRef = useRef(null);

  useEffect(() => {
    if (imageRef.current) {
      parallaxRef.current = new SimpleParallax(imageRef.current, {
        scale,
        orientation,
        overflow,
        delay,
        transition,
        maxTransition
      });
    }

    return () => {
      if (parallaxRef.current) {
        parallaxRef.current.destroy();
      }
    };
  }, [scale, orientation, overflow, delay, transition, maxTransition]);

  return (
    <img
      ref={imageRef}
      src={src}
      alt={alt}
      className={className}
    />
  );
}
