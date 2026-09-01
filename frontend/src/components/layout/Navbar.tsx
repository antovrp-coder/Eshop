import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ShoppingBag, Search, Compass, PackageCheck, User } from 'lucide-react';
import { Cart } from '../../services/graphqlClient';

interface NavbarProps {
  cart: Cart | null;
  onOpenCart: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ cart, onOpenCart }) => {
  const navigate = useNavigate();
  const totalItems = cart?.items.reduce((sum, it) => sum + it.quantity, 0) || 0;

  return (
    <header className="glass-panel" style={{ position: 'sticky', top: 0, zIndex: 50, borderBottom: '1px solid var(--border-subtle)' }}>
      <div className="container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: '4.5rem' }}>
        {/* Brand Logo */}
        <Link to="/" style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{
            width: '2.25rem',
            height: '2.25rem',
            borderRadius: 'var(--radius-sm)',
            background: 'linear-gradient(135deg, var(--accent-gold) 0%, #a8793b 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 800,
            color: '#000',
            fontSize: '1.2rem',
          }}>
            S
          </div>
          <div>
            <span className="serif" style={{ fontSize: '1.4rem', fontWeight: 700, letterSpacing: '-0.02em' }}>StyleHome</span>
            <span style={{ fontSize: '0.65rem', display: 'block', color: 'var(--accent-gold)', letterSpacing: '0.15em', textTransform: 'uppercase', marginTop: '-4px' }}>
              Atelier Living
            </span>
          </div>
        </Link>

        {/* Navigation Links */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <Link to="/search" style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', fontWeight: 500, display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <Compass size={18} /> Explore Catalog
          </Link>
          <Link to="/search?category=living-room" style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', fontWeight: 500 }}>
            Living Room
          </Link>
          <Link to="/search?category=lighting" style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', fontWeight: 500 }}>
            Lighting
          </Link>
          <Link to="/search?category=workspace" style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', fontWeight: 500 }}>
            Workspace
          </Link>
          <Link to="/orders" style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', fontWeight: 500, display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <PackageCheck size={18} /> My Orders
          </Link>
        </nav>

        {/* Action Controls */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <button
            onClick={() => navigate('/search')}
            style={{
              padding: '0.6rem',
              borderRadius: 'var(--radius-full)',
              background: 'var(--bg-surface-elevated)',
              color: 'var(--text-secondary)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
            title="Search Products (Elasticsearch)"
          >
            <Search size={19} />
          </button>

          <button
            onClick={onOpenCart}
            style={{
              position: 'relative',
              padding: '0.6rem 1rem',
              borderRadius: 'var(--radius-full)',
              background: 'var(--bg-surface-elevated)',
              color: 'var(--text-primary)',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              border: '1px solid var(--border-subtle)',
            }}
          >
            <ShoppingBag size={19} color="var(--accent-gold)" />
            <span style={{ fontSize: '0.9rem', fontWeight: 600 }}>Cart</span>
            {totalItems > 0 && (
              <span style={{
                background: 'var(--accent-gold)',
                color: '#000',
                fontSize: '0.75rem',
                fontWeight: 700,
                padding: '0.1rem 0.45rem',
                borderRadius: 'var(--radius-full)',
              }}>
                {totalItems}
              </span>
            )}
          </button>
        </div>
      </div>
    </header>
  );
};
