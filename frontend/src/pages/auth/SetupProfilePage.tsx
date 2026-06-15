import { Check, PawPrint } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
import { useAuth } from '../../app/useAuth';
import { completeProfile } from '../../shared/api/auth';
import { clearProfileSetupToken, getProfileSetupToken } from '../../shared/lib/profile-setup-token';

const petStages = [
  { value: 'planning', label: '准备养宠' },
  { value: 'comparing', label: '正在比较品种' },
  { value: 'already_have', label: '已经有宠物' },
  { value: 'helping_family', label: '帮家人做选择' },
];

const petTypes = [
  { value: 'dog', label: '小狗' },
  { value: 'cat', label: '小猫' },
  { value: 'small_pet', label: '小型宠物' },
  { value: 'not_sure', label: '还不确定' },
];

const constraints = [
  { value: 'allergy', label: '过敏顾虑' },
  { value: 'rental', label: '租房限制' },
  { value: 'budget', label: '预算有限' },
  { value: 'travel', label: '经常出差' },
];

export function SetupProfilePage() {
  const navigate = useNavigate();
  const auth = useAuth();
  const [profileSetupToken] = useState(() => getProfileSetupToken());
  const [nickname, setNickname] = useState('');
  const [petStage, setPetStage] = useState('planning');
  const [interestedPetTypes, setInterestedPetTypes] = useState<string[]>(['dog']);
  const [petExperience, setPetExperience] = useState('');
  const [dailyCompanyTime, setDailyCompanyTime] = useState('');
  const [livingSituation, setLivingSituation] = useState('');
  const [petConstraints, setPetConstraints] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!profileSetupToken) {
      navigate('/auth/login', { replace: true });
    }
  }, [navigate, profileSetupToken]);

  function toggle(list: string[], value: string, setter: (next: string[]) => void) {
    setter(list.includes(value) ? list.filter((item) => item !== value) : [...list, value]);
  }

  async function handleSubmit() {
    setSubmitting(true);
    setError('');
    try {
      const data = await completeProfile({
        profileSetupToken,
        nickname,
        petStage,
        interestedPetTypes,
        petExperience,
        dailyCompanyTime,
        livingSituation,
        petConstraints,
      });
      clearProfileSetupToken();
      await auth.loginWithAuth(data);
      navigate('/');
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-page">
      <AuthAnimalStage mode="account" />
      <section className="auth-panel auth-panel--wide" aria-labelledby="setup-title">
        <div className="auth-panel__header">
          <p className="eyebrow">基础资料</p>
          <h1 id="setup-title">让 Two-To 先认识你</h1>
          <p>这些信息会用于后续适配测评和推荐排序。</p>
        </div>

        <form className="form-stack" onSubmit={(event) => event.preventDefault()}>
          <label className="field">
            <span>昵称</span>
            <div className="field__control">
              <PawPrint size={17} />
              <input value={nickname} onChange={(event) => setNickname(event.target.value)} placeholder="2-20 个字符" />
            </div>
          </label>

          <div className="field">
            <span>养宠阶段</span>
            <div className="choice-grid">
              {petStages.map((item) => (
                <button className={petStage === item.value ? 'choice choice--active' : 'choice'} key={item.value} onClick={() => setPetStage(item.value)} type="button">
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          <div className="field">
            <span>感兴趣宠物</span>
            <div className="choice-grid">
              {petTypes.map((item) => (
                <button
                  className={interestedPetTypes.includes(item.value) ? 'choice choice--active' : 'choice'}
                  key={item.value}
                  onClick={() => toggle(interestedPetTypes, item.value, setInterestedPetTypes)}
                  type="button"
                >
                  {interestedPetTypes.includes(item.value) ? <Check size={15} /> : null}
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          <div className="form-grid">
            <label className="field">
              <span>养宠经验</span>
              <select value={petExperience} onChange={(event) => setPetExperience(event.target.value)}>
                <option value="">暂不填写</option>
                <option value="newbie">第一次养</option>
                <option value="had_before">以前养过</option>
                <option value="experienced">比较熟悉</option>
              </select>
            </label>
            <label className="field">
              <span>每日陪伴时间</span>
              <select value={dailyCompanyTime} onChange={(event) => setDailyCompanyTime(event.target.value)}>
                <option value="">暂不填写</option>
                <option value="lt_1h">少于 1 小时</option>
                <option value="1_3h">1-3 小时</option>
                <option value="3_6h">3-6 小时</option>
                <option value="gt_6h">6 小时以上</option>
              </select>
            </label>
          </div>

          <label className="field">
            <span>居住情况</span>
            <select value={livingSituation} onChange={(event) => setLivingSituation(event.target.value)}>
              <option value="">暂不填写</option>
              <option value="alone">独居</option>
              <option value="family">和家人同住</option>
              <option value="roommate">和室友同住</option>
              <option value="with_child">家里有小朋友</option>
            </select>
          </label>

          <div className="field">
            <span>养宠限制</span>
            <div className="choice-grid">
              {constraints.map((item) => (
                <button
                  className={petConstraints.includes(item.value) ? 'choice choice--active' : 'choice'}
                  key={item.value}
                  onClick={() => toggle(petConstraints, item.value, setPetConstraints)}
                  type="button"
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {error ? <p className="form-error">{error}</p> : null}
          <button className="primary-button" disabled={submitting} onClick={handleSubmit} type="button">
            <Check size={18} />
            {submitting ? '保存中' : '完成并进入'}
          </button>
        </form>
      </section>
    </main>
  );
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return '操作失败，请稍后再试';
}
