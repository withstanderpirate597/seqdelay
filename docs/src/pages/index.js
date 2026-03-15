import React from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
function Hero() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header style={{ padding: '5rem 0 4rem', textAlign: 'center', background: 'linear-gradient(135deg, #0f172a 0%, #1e3a8a 50%, #2563eb 100%)', color: '#fff' }}>
      <div className="container">
        <p style={{ fontSize: '0.875rem', textTransform: 'uppercase', letterSpacing: '3px', opacity: 0.7, marginBottom: '1rem' }}>Delay Queue</p>
        <h1 style={{ fontSize: '4rem', fontWeight: 800, marginBottom: '1rem', letterSpacing: '-2px' }}>{siteConfig.title}</h1>
        <p style={{ fontSize: '1.3rem', opacity: 0.9, maxWidth: '600px', margin: '0 auto 2.5rem', lineHeight: 1.6 }}>{siteConfig.tagline}</p>
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link className="button button--lg" to="/docs/getting-started" style={{ background: '#fff', color: '#1e3a8a', fontWeight: 700, border: 'none' }}>Get Started</Link>
          <Link className="button button--lg" href="https://github.com/gocronx/seqdelay" style={{ background: 'transparent', color: '#fff', border: '2px solid rgba(255,255,255,0.4)' }}>View on GitHub</Link>
        </div>
      </div>
    </header>
  );
}
function Features() {
  const features = [
    { title: 'Time Wheel Engine', desc: 'Powered by seqflow Disruptor. O(1) add/fire. 4.4M tasks/sec. 1ms precision.', color: '#2563eb' },
    { title: 'Redis Persistence', desc: 'Lua scripts for atomic state transitions. Standalone, Sentinel, or Cluster. Full recovery on restart.', color: '#3b82f6' },
    { title: 'Dual Mode', desc: 'Embed as Go library with callbacks, or deploy as standalone HTTP service for any language.', color: '#60a5fa' },
  ];
  return (
    <section style={{ padding: '4rem 0', background: '#f8fafc' }}>
      <div className="container">
        <div className="row">
          {features.map((f, i) => (
            <div key={i} className="col col--4" style={{ marginBottom: '2rem' }}>
              <div style={{ padding: '2rem', borderRadius: '12px', background: '#fff', border: '1px solid #e2e8f0', height: '100%', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
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
  const cases = ['Order auto-cancel (15 min unpaid)', 'Payment callback retry', 'Membership expiry reminder', 'Coupon expiration', 'Scheduled push notifications', 'Rate limit cooldown'];
  return (
    <section style={{ padding: '4rem 0' }}>
      <div className="container" style={{ maxWidth: '700px' }}>
        <h2 style={{ textAlign: 'center', marginBottom: '2rem' }}>Use Cases</h2>
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
    <Layout title="Home" description="High-performance delay queue powered by seqflow + Redis">
      <Hero /><Features /><UseCases />
    </Layout>
  );
}
