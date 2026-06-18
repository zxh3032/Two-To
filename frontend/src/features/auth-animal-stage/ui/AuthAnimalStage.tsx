import { Sparkles } from 'lucide-react';
import { useMemo } from 'react';

type AnimalMode = 'idle' | 'account' | 'password' | 'secret';

interface AuthAnimalStageProps {
  mode: AnimalMode;
}

// AuthAnimalStage 承载登录页左侧品牌区，用油画场景表达 Two-To 的宠物陪伴气质。
export function AuthAnimalStage({ mode }: AuthAnimalStageProps) {
  const modeClass = useMemo(() => `brand-stage--${mode}`, [mode]);

  return (
    <section className={`brand-stage ${modeClass}`} aria-label="Two-To 品牌介绍">
      <div className="brand-stage__art" aria-hidden="true">
        <picture>
          <source media="(max-width: 620px)" srcSet="/auth-companion-oil-portrait.png?v=20260618d" />
          <source media="(max-width: 1100px)" srcSet="/auth-companion-oil-wide.png?v=20260618d" />
          <img src="/auth-companion-oil-web.png?v=20260618d" alt="" />
        </picture>
      </div>
      <div className="brand-stage__paint" aria-hidden="true" />
      <div className="brand-stage__privacy" aria-hidden="true" />

      <div className="brand-stage__copy">
        <div className="brand-stage__mark">
          <img src="/two-to-mark.svg" alt="" />
          <span>Two-To</span>
        </div>
        <p className="brand-stage__eyebrow">
          <Sparkles size={15} />
          宠物适配与长期照料
        </p>
        <h1>
          <span>相逢以后，</span>
          <span>把心动照料成日常。</span>
        </h1>
        <p>登录后继续测评、档案和照护记录，让每一次相处都有脉络。</p>
      </div>
    </section>
  );
}
