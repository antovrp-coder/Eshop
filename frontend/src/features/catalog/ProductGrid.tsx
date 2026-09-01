import React from 'react';
import { Link } from 'react-router-dom';
import { Star, ShoppingBag } from 'lucide-react';
import { Product, formatPrice } from '../../services/graphqlClient';

interface ProductGridProps {
  products: Product[];
  onAddToCart?: (variantId: string) => void;
}

export const ProductGrid: React.FC<ProductGridProps> = ({ products, onAddToCart }) => {
  if (!products || products.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: '5rem 1rem', color: 'var(--text-secondary)' }}>
        <p style={{ fontSize: '1.1rem' }}>No products match your current selection.</p>
      </div>
    );
  }

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '2rem' }}>
      {products.map((product) => {
        const defaultVariant = product.variants[0];
        return (
          <div
            key={product.id}
            style={{
              background: 'var(--bg-surface)',
              borderRadius: 'var(--radius-lg)',
              overflow: 'hidden',
              border: '1px solid var(--border-subtle)',
              display: 'flex',
              flexDirection: 'column',
              transition: 'var(--transition-smooth)',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px)';
              e.currentTarget.style.boxShadow = 'var(--shadow-elevated)';
              e.currentTarget.style.borderColor = 'var(--border-hover)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = 'none';
              e.currentTarget.style.borderColor = 'var(--border-subtle)';
            }}
          >
            {/* Product Image Box */}
            <Link to={`/product/${product.slug}`} style={{ position: 'relative', height: '260px', overflow: 'hidden', display: 'block' }}>
              <img
                src={product.featured_image}
                alt={product.title}
                style={{ width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.4s ease' }}
                onMouseEnter={(e) => (e.currentTarget.style.transform = 'scale(1.05)')}
                onMouseLeave={(e) => (e.currentTarget.style.transform = 'scale(1)')}
              />
              <div style={{ position: 'absolute', top: '1rem', left: '1rem', display: 'flex', gap: '0.4rem' }}>
                {product.is_featured && <span className="badge badge-gold">Featured</span>}
              </div>
            </Link>

            {/* Product Information */}
            <div style={{ padding: '1.25rem', flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
              <div>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.4rem' }}>
                  <span style={{ fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--text-muted)' }}>
                    {product.brand}
                  </span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.8rem', color: 'var(--accent-gold)' }}>
                    <Star size={13} fill="currentColor" />
                    <span>{product.average_rating}</span>
                  </div>
                </div>

                <Link to={`/product/${product.slug}`}>
                  <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-primary)', lineHeight: 1.4, marginBottom: '0.5rem' }}>
                    {product.title}
                  </h3>
                </Link>

                <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', marginBottom: '1rem' }}>
                  {product.description}
                </p>
              </div>

              {/* Price & Quick Add */}
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)' }}>
                <div>
                  <span style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--accent-gold)' }}>
                    {formatPrice(product.base_price_cents)}
                  </span>
                </div>

                {onAddToCart && defaultVariant && (
                  <button
                    onClick={() => onAddToCart(defaultVariant.id)}
                    className="btn-secondary"
                    style={{ padding: '0.5rem 0.85rem', fontSize: '0.8rem' }}
                    title="Quick Add to Cart"
                  >
                    <ShoppingBag size={15} /> Add
                  </button>
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
