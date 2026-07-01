import { CheckCircle2, Route, ShieldCheck, Sparkles } from 'lucide-react';

type AnimalMode = 'idle' | 'account' | 'password' | 'secret';

interface AuthAnimalStageProps {
  mode: AnimalMode;
}

const stageItems = [
  { icon: CheckCircle2, title: '先判断', text: '把养宠前的状态、时间和限制梳理清楚。' },
  { icon: Route, title: '再选择', text: '用适配逻辑连接品种、性格和照护成本。' },
  { icon: ShieldCheck, title: '长期记录', text: '把档案、安全和照护节点留在同一处。' },
];

// AuthAnimalStage 承载登录页左侧品牌区，用抽象路径表达 Two-To 的适配与陪伴。
export function AuthAnimalStage({ mode }: AuthAnimalStageProps) {
  return (
    <section className={`brand-stage brand-stage--${mode}`} aria-label="Two-To 品牌介绍">
      <div className="brand-stage__copy">
        <div className="brand-stage__mark">
          <img src="/two-to-mark.png" alt="" />
          <span>Two-To</span>
        </div>
        <p className="brand-stage__eyebrow">
          <Sparkles size={15} />
          宠物适配与长期照料
        </p>
        <h1>
          <span>先判断，</span>
          <span>再开始认真陪伴。</span>
        </h1>
        <p>Two-To 把养前评估、账号资料和照护记录放进一条清晰路径里。</p>
      </div>

      <div className="brand-stage__path" aria-hidden="true">
        <span />
        <span />
        <span />
      </div>

      <div className="brand-stage__list">
        {stageItems.map((item) => (
          <div className="brand-stage__item" key={item.title}>
            <item.icon size={17} />
            <div>
              <strong>{item.title}</strong>
              <span>{item.text}</span>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
