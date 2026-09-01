import React from 'react';
import { ShieldCheck, Truck, RotateCcw, Award } from 'lucide-react';

export const Footer: React.FC = () => {
  return (
    <footer style={{ background: 'var(--bg-surface)', borderTop: '1px solid var(--border-subtle)', marginTop: '6rem' }}>
      {/* Value Propositions */}
      <div className="container" style={{ padding: '3.5rem 1.5rem', borderBottom: '1px solid var(--border-subtle)' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '2rem' }}>
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <div style={{ padding: '0.75rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-md)', color: 'var(--accent-gold)' }}>
              <Award size={24} />
            </div>
            <div>
              <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>Master Craftsmanship</h4>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>FSC-certified hardwoods and artisan glazes</p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <div style={{ padding: '0.75rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-md)', color: 'var(--accent-gold)' }}>
              <Truck size={24} />
            </div>
            <div>
              <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>White-Glove Delivery</h4>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Complimentary assembly on orders $100+</p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <div style={{ padding: '0.75rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-md)', color: 'var(--accent-gold)' }}>
              <RotateCcw size={24} />
            </div>
            <div>
              <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>30-Day Home Trial</h4>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Hassle-free returns and exchanges</p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <div style={{ padding: '0.75rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-md)', color: 'var(--accent-gold)' }}>
              <ShieldCheck size={24} />
            </div>
            <div>
              <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>PCI-DSS Bank Grade</h4>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>End-to-end tokenized checkout</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Footer Links */}
      <div className="container" style={{ padding: '4rem 1.5rem', display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: '2rem' }}>
        <div style={{ maxWidth: '360px' }}>
          <div className="serif" style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.75rem' }}>StyleHome</div>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', lineHeight: 1.7 }}>
            Elevating modern living spaces through sculptural silhouettes, sustainable materials, and uncompromising architecture.
          </p>
        </div>

        <div>
          <h5 style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '1rem' }}>
            Collections
          </h5>
          <ul style={{ listStyle: 'none', display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            <li>Living Room Seating</li>
            <li>Architectural Dining</li>
            <li>Lumiere Scandinavian Lighting</li>
            <li>Handmade Ceramic Vases</li>
          </ul>
        </div>

        <div>
          <h5 style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '1rem' }}>
            Technology & Arch
          </h5>
          <ul style={{ listStyle: 'none', display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            <li>Go Standard Library (Zero Framework)</li>
            <li>gRPC Microservices Mesh</li>
            <li>Elasticsearch Full-Text Engine</li>
            <li>PostgreSQL Eshop Cluster (CQRS)</li>
          </ul>
        </div>
      </div>

      <div style={{ borderTop: '1px solid var(--border-subtle)', padding: '1.5rem 0', textAlign: 'center', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
        © 2026 StyleHome. Built with Go (net/http + gRPC) and React TypeScript. All rights reserved.
      </div>
    </footer>
  );
};
