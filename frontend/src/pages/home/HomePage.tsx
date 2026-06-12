import { ProjectSlogan } from '../../features/project-slogan/ui/ProjectSlogan';

const moduleCards = [
  { title: '适配测评', status: '即将开放', description: '先判断现在的生活状态，到底适不适合开始养。' },
  { title: '品种解析', status: '即将开放', description: '再理解不同宠物的性格、照护难度和适合场景。' },
  { title: '宠物档案', status: '即将开放', description: '养了之后，把资料、成长节点和陪伴回忆长期记录下来。' },
  { title: '健康照料', status: '即将开放', description: '把疫苗、驱虫、体重、用药和就诊提醒持续管起来。' },
];

// HomePage 负责首页首屏和功能入口编排，不直接展示接口调试信息。
// 请求参数：无；返回值：面向用户的首页 React 节点。
export function HomePage() {
  return (
    <section className="page-stack home-page">
      <ProjectSlogan />

      {/* 后端联通状态卡片先不挂到首页，避免用户视角出现调试型信息。 */}
      {/* <ApiHealthCheck /> */}

      <section className="home-section">
        <div>
          <p className="eyebrow">首期能力</p>
          <h2>从想养，到养好</h2>
          <p>从养前评估、品种选择到长期照护，让每一次陪伴都有清晰依据和持续记录。</p>
        </div>

        <div className="module-grid" aria-label="首期功能模块">
          {moduleCards.map((item) => (
            <article className="module-card" key={item.title}>
              <div className="module-card__top">
                <h3>{item.title}</h3>
                <span>{item.status}</span>
              </div>
              <p>{item.description}</p>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
}
