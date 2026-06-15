import { useMemo, useRef, useState } from 'react';
import type { CSSProperties, PointerEvent } from 'react';

type AnimalMode = 'idle' | 'account' | 'password' | 'secret';
type AnimalTone = 'cream-dog' | 'brown-dog' | 'black-cat' | 'cheese-cat' | 'maine-cat';

interface AuthAnimalStageProps {
  mode: AnimalMode;
}

const animals: Array<{ key: string; label: string; tone: AnimalTone; left: number; y: number; scale: number; z: number }> = [
  { key: 'pomeranian', label: '白色博美', tone: 'cream-dog', left: 13, y: 12, scale: 0.96, z: 3 },
  { key: 'teddy', label: '棕色泰迪', tone: 'brown-dog', left: 30, y: 2, scale: 1.02, z: 5 },
  { key: 'devon', label: '黑色德文', tone: 'black-cat', left: 49, y: 18, scale: 0.94, z: 7 },
  { key: 'cheese', label: '黑白美短起司', tone: 'cheese-cat', left: 68, y: 6, scale: 1, z: 6 },
  { key: 'maine', label: '白色缅因', tone: 'maine-cat', left: 86, y: 16, scale: 1.04, z: 2 },
];

// AuthAnimalStage 负责登录页品牌区的互动小动物舞台。
export function AuthAnimalStage({ mode }: AuthAnimalStageProps) {
  const stageRef = useRef<HTMLDivElement | null>(null);
  const [eye, setEye] = useState({ x: 0, y: 0 });
  const modeClass = useMemo(() => `animal-stage--${mode}`, [mode]);

  function handlePointerMove(event: PointerEvent<HTMLDivElement>) {
    const rect = stageRef.current?.getBoundingClientRect();
    if (!rect) {
      return;
    }
    const dx = (event.clientX - rect.left) / rect.width - 0.5;
    const dy = (event.clientY - rect.top) / rect.height - 0.5;
    setEye({
      x: Math.max(-1, Math.min(1, dx * 2)),
      y: Math.max(-1, Math.min(1, dy * 2)),
    });
  }

  const style = {
    '--eye-x': `${eye.x * 5}px`,
    '--eye-y': `${eye.y * 4}px`,
  } as CSSProperties;

  return (
    <section className={`animal-stage ${modeClass}`} onPointerMove={handlePointerMove} ref={stageRef} style={style}>
      <div className="animal-stage__sky" />
      <div className="animal-stage__rail" />
      <div className="animal-pack" aria-label="Two-To 小动物伙伴">
        {animals.map((animal, index) => (
          <div
            className={`animal animal--${animal.tone}`}
            key={animal.key}
            style={
              {
                '--animal-left': `${animal.left}%`,
                '--animal-y': `${animal.y}px`,
                '--animal-scale': animal.scale,
                '--animal-z': animal.z,
                '--animal-delay': `${index * 80}ms`,
              } as CSSProperties
            }
            title={animal.label}
          >
            <AnimalIllustration secret={mode === 'secret'} tone={animal.tone} />
          </div>
        ))}
      </div>
      <div className="animal-stage__copy">
        <p className="eyebrow">Two-To</p>
        <h1>两两相逢，先认出彼此</h1>
      </div>
    </section>
  );
}

function AnimalIllustration({ secret, tone }: { secret: boolean; tone: AnimalTone }) {
  const frontStyle = { display: secret ? 'none' : 'inline', opacity: secret ? 0 : 1 } as CSSProperties;
  const backStyle = { display: secret ? 'inline' : 'none', opacity: secret ? 1 : 0 } as CSSProperties;

  if (tone === 'cream-dog') {
    return (
      <svg className="animal__svg" viewBox="0 0 180 220" aria-hidden="true" focusable="false">
        <ellipse className="animal__ground" cx="90" cy="205" rx="62" ry="12" />
        <g className="animal__back" style={backStyle}>
          <path d="M43 188c-8-28-3-76 22-99 22-21 59-21 82 1 24 23 29 72 20 98z" fill="#fff4dc" />
          <path d="M122 95c22-12 35 4 25 22-8 15-30 12-35-2" fill="#ffe1ba" />
          <circle cx="94" cy="103" r="23" fill="#ffe8bd" />
        </g>
        <g className="animal__front" style={frontStyle}>
          <path d="M40 190c-10-30-2-79 25-103 25-22 65-22 90 2 25 24 32 72 23 101z" fill="#fff5df" />
          <path d="M47 101c-22-16-12-42 12-34 15 5 22 22 16 36z" fill="#ffe5bd" />
          <path d="M134 101c21-17 11-43-13-35-15 5-22 22-16 37z" fill="#ffe5bd" />
          <path d="M49 141c-18-8-24-23-14-35 10-13 31-7 38 11" fill="#fff9eb" />
          <path d="M132 116c9-18 30-24 39-11 10 13 3 29-16 37" fill="#fff9eb" />
          <circle cx="91" cy="116" r="45" fill="#fffaf0" />
          <path d="M63 157c16 16 43 17 58 0" className="animal__mouth-line" />
          <g className="animal__eyes">
            <circle cx="73" cy="105" r="9" fill="#fffdf8" />
            <circle cx="73" cy="105" r="4.6" fill="#3a2418" />
            <circle cx="111" cy="105" r="9" fill="#fffdf8" />
            <circle cx="111" cy="105" r="4.6" fill="#3a2418" />
          </g>
          <ellipse cx="92" cy="127" rx="10" ry="7" fill="#3a2418" />
          <circle className="animal__cheek" cx="58" cy="130" r="7" />
          <circle className="animal__cheek" cx="126" cy="130" r="7" />
          <path className="animal__front-paw" d="M60 186c5-17 28-17 32 1" fill="#ffe9c8" />
          <path className="animal__front-paw" d="M93 187c5-18 29-18 33 0" fill="#ffe9c8" />
        </g>
      </svg>
    );
  }

  if (tone === 'brown-dog') {
    return (
      <svg className="animal__svg" viewBox="0 0 180 220" aria-hidden="true" focusable="false">
        <ellipse className="animal__ground" cx="92" cy="205" rx="64" ry="12" />
        <g className="animal__back" style={backStyle}>
          <path d="M43 190c-8-37 3-92 47-104 44-11 76 22 78 104z" fill="#9c5b35" />
          <path d="M131 103c20-17 39 0 27 20-10 16-33 11-37-5" fill="#6f3d27" />
          <circle cx="91" cy="112" r="21" fill="#c68458" />
        </g>
        <g className="animal__front" style={frontStyle}>
          <path d="M41 190c-9-39 4-94 48-107 45-12 78 22 80 107z" fill="#a8643d" />
          <path d="M52 97c-25-21-18-55 8-49 18 4 26 25 18 53z" fill="#7c4328" />
          <path d="M129 98c24-22 17-56-9-50-18 4-25 25-17 53z" fill="#7c4328" />
          <circle cx="90" cy="118" r="46" fill="#a8643d" />
          <ellipse cx="91" cy="139" rx="28" ry="22" fill="#e2b38d" />
          <g className="animal__eyes">
            <circle cx="70" cy="107" r="9" fill="#fffdf8" />
            <circle cx="70" cy="107" r="4.4" fill="#3a2418" />
            <circle cx="111" cy="107" r="9" fill="#fffdf8" />
            <circle cx="111" cy="107" r="4.4" fill="#3a2418" />
          </g>
          <ellipse cx="91" cy="135" rx="10" ry="7" fill="#3a2418" />
          <path d="M81 151c7 8 15 8 22 0" className="animal__mouth-line" />
          <path className="animal__front-paw" d="M53 186c6-19 31-20 37 0" fill="#d89a70" />
          <path className="animal__front-paw" d="M93 186c6-20 32-20 38 0" fill="#d89a70" />
        </g>
      </svg>
    );
  }

  if (tone === 'black-cat') {
    return (
      <svg className="animal__svg" viewBox="0 0 180 220" aria-hidden="true" focusable="false">
        <ellipse className="animal__ground" cx="90" cy="205" rx="58" ry="12" />
        <g className="animal__back" style={backStyle}>
          <path d="M45 190c-5-39 8-86 45-101 38 15 51 62 45 101z" fill="#1f2723" />
          <path d="M126 152c35-8 45 21 20 34-17 9-38-2-39-21" fill="#111813" />
        </g>
        <g className="animal__front" style={frontStyle}>
          <path d="M43 190c-6-42 9-89 47-105 39 16 53 63 47 105z" fill="#202824" />
          <path d="M57 94 69 38l31 39z" fill="#111813" />
          <path d="m120 94-10-56-33 39z" fill="#111813" />
          <path d="M65 78 72 52l13 22z" fill="#3f2922" />
          <path d="m112 78-5-26-15 22z" fill="#3f2922" />
          <circle cx="90" cy="124" r="42" fill="#202824" />
          <g className="animal__eyes animal__eyes--gold">
            <ellipse cx="73" cy="112" rx="11" ry="13" fill="#ffd166" />
            <ellipse cx="73" cy="112" rx="3.5" ry="8" fill="#15201b" />
            <ellipse cx="108" cy="112" rx="11" ry="13" fill="#ffd166" />
            <ellipse cx="108" cy="112" rx="3.5" ry="8" fill="#15201b" />
          </g>
          <ellipse cx="91" cy="139" rx="11" ry="8" fill="#d9c7a7" />
          <ellipse cx="91" cy="133" rx="9" ry="6" fill="#111813" />
          <path d="M53 138h25M52 149h26M104 138h25M104 149h27" className="animal__whiskers" />
          <path className="animal__front-paw" d="M59 188c5-19 28-19 33 0" fill="#445048" />
          <path className="animal__front-paw" d="M91 188c5-19 28-19 33 0" fill="#445048" />
        </g>
      </svg>
    );
  }

  if (tone === 'cheese-cat') {
    return (
      <svg className="animal__svg" viewBox="0 0 180 220" aria-hidden="true" focusable="false">
        <ellipse className="animal__ground" cx="90" cy="205" rx="62" ry="12" />
        <g className="animal__back" style={backStyle}>
          <path d="M42 190c-8-42 10-88 49-101 38 14 55 60 47 101z" fill="#fff7e8" />
          <path d="M114 109c18 2 34 12 40 28-16 13-38 9-51-6z" fill="#202427" />
        </g>
        <g className="animal__front" style={frontStyle}>
          <path d="M40 190c-8-42 11-89 50-104 39 15 57 62 49 104z" fill="#fff8ef" />
          <path d="M57 97 69 42l31 40z" fill="#fff8ef" />
          <path d="m121 96-10-54-33 39z" fill="#202427" />
          <path d="M91 78c28 0 50 19 50 48 0 31-21 52-50 52s-51-21-51-52c0-29 22-48 51-48z" fill="#fffaf1" />
          <path d="M93 79c24 3 44 22 44 48 0 13-4 24-12 33-18-10-28-33-24-80z" fill="#202427" />
          <g className="animal__eyes animal__eyes--gold">
            <circle cx="72" cy="112" r="10" fill="#ffd166" />
            <circle cx="72" cy="112" r="4" fill="#3a2418" />
            <circle cx="109" cy="112" r="10" fill="#ffd166" />
            <circle cx="109" cy="112" r="4" fill="#3a2418" />
          </g>
          <ellipse cx="90" cy="134" rx="9" ry="6" fill="#3a2418" />
          <path d="M55 140h23M54 151h24M103 140h24M103 151h25" className="animal__whiskers" />
          <path className="animal__front-paw" d="M58 187c6-18 30-18 35 0" fill="#f1e3ce" />
          <path className="animal__front-paw" d="M91 187c6-18 30-18 35 0" fill="#f1e3ce" />
        </g>
      </svg>
    );
  }

  return (
    <svg className="animal__svg" viewBox="0 0 180 220" aria-hidden="true" focusable="false">
      <ellipse className="animal__ground" cx="90" cy="205" rx="68" ry="12" />
      <g className="animal__back" style={backStyle}>
        <path d="M35 190c-7-47 16-100 58-110 42 10 65 63 58 110z" fill="#fbf3e6" />
        <path d="M128 124c30-15 51 12 33 34-15 18-43 9-47-15" fill="#efe3d1" />
        <path d="M65 98c17-12 40-12 57 0" className="animal__fur-line" />
      </g>
      <g className="animal__front" style={frontStyle}>
        <path d="M32 190c-8-49 17-103 60-112 43 9 68 63 60 112z" fill="#fbf5ea" />
        <path d="M57 96 68 34l31 45z" fill="#fbf5ea" />
        <path d="m124 96-11-62-32 45z" fill="#fbf5ea" />
        <path d="M71 65 75 46l12 20zM110 65l-4-19-12 20z" fill="#f1d8c8" />
        <path d="M50 121c-6 23 4 51 21 62M130 121c7 23-3 51-20 62" className="animal__fur-line" />
        <circle cx="91" cy="122" r="48" fill="#fffaf1" />
        <g className="animal__eyes">
          <circle cx="72" cy="111" r="9" fill="#fffdf8" />
          <circle cx="72" cy="111" r="4.5" fill="#3a2418" />
          <circle cx="110" cy="111" r="9" fill="#fffdf8" />
          <circle cx="110" cy="111" r="4.5" fill="#3a2418" />
        </g>
        <ellipse cx="91" cy="135" rx="9" ry="6" fill="#3a2418" />
        <path d="M82 150c7 7 14 7 21 0" className="animal__mouth-line" />
        <circle className="animal__cheek" cx="58" cy="132" r="7" />
        <circle className="animal__cheek" cx="124" cy="132" r="7" />
        <path className="animal__front-paw" d="M55 187c6-20 32-20 38 0" fill="#efe3d1" />
        <path className="animal__front-paw" d="M91 187c6-20 33-20 39 0" fill="#efe3d1" />
      </g>
    </svg>
  );
}
