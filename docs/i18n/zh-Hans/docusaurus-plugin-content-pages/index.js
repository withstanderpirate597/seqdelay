import React from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

function Hero() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header style={{
      padding: '5rem 0 4rem',
      textAlign: 'center',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e3a8a 50%, #2563eb 100%)',
      color: '#fff',
    }}>
      <div className="container">
        <p style={{
          fontSize: '0.875rem',
          textTransform: 'uppercase',
          letterSpacing: '3px',
          opacity: 0.7,
          marginBottom: '1rem',
        }}>
          延迟队列
        </p>
        <h1 style={{
          fontSize: '4rem',
          fontWeight: 800,
          marginBottom: '1rem',
          letterSpacing: '-2px',
        }}>
          {siteConfig.title}
        </h1>
        <p style={{
          fontSize: '1.3rem',
          opacity: 0.9,
          maxWidth: '600px',
          margin: '0 auto 2.5rem',
          lineHeight: 1.6,
        }}>
          基于 seqflow 时间轮 + Redis 的高性能延迟队列
        </p>
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link className="button button--lg" to="/zh-Hans/docs/getting-started"
            style={{ background: '#fff', color: '#1e3a8a', fontWeight: 700, border: 'none' }}>
            快速开始
          </Link>
          <Link className="button button--lg"
            href="https://github.com/gocronx/seqdelay"
            style={{ background: 'transparent', color: '#fff', border: '2px solid rgba(255,255,255,0.4)' }}>
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

function Features() {
  const features = [
    { title: '时间轮引擎', desc: '基于 seqflow Disruptor。O(1) 添加/触发。440 万 tasks/sec。1ms 精度。', color: '#2563eb' },
    { title: 'Redis 持久化', desc: 'Lua 脚本原子状态转换。支持单机、Sentinel、Cluster。重启后完整恢复。', color: '#3b82f6' },
    { title: '双模式', desc: '嵌入 Go 服务用回调，或独立部署为 HTTP 服务供任何语言调用。', color: '#60a5fa' },
  ];
  return (
    <section style={{ padding: '4rem 0', background: '#f8fafc' }}>
      <div className="container">
        <div className="row">
          {features.map((f, i) => (
            <div key={i} className="col col--4" style={{ marginBottom: '2rem' }}>
              <div style={{
                padding: '2rem', borderRadius: '12px', background: '#fff',
                border: '1px solid #e2e8f0', height: '100%', boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
              }}>
                <div style={{ width: 8, height: 8, borderRadius: '50%', background: f.color, marginBottom: '1rem' }} />
                <h3 style={{ fontSize: '1.15rem', marginBottom: '0.75rem' }}>{f.title}</h3>
                <p style={{ color: '#64748b', lineHeight: 1.6, marginBottom: 0 }}>{f.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function UseCases() {
  const cases = ['订单 15 分钟未支付自动取消', '支付回调递增重试', '会员到期提醒', '优惠券过期', '定时推送', '限流冷却'];
  return (
    <section style={{ padding: '4rem 0' }}>
      <div className="container" style={{ maxWidth: '700px', margin: '0 auto' }}>
        <h2 style={{ textAlign: 'center', marginBottom: '2rem' }}>应用场景</h2>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
          {cases.map((c, i) => (
            <div key={i} style={{ padding: '1rem 1.25rem', background: '#f1f5f9', borderRadius: '8px', borderLeft: '3px solid #2563eb', fontSize: '0.95rem' }}>{c}</div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  return (
    <Layout title="首页" description="基于 seqflow 时间轮 + Redis 的高性能延迟队列">
      <Hero /><Features /><UseCases />
    </Layout>
  );
}
