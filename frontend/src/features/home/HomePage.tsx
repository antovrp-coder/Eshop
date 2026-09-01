import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, Sparkles, Compass } from 'lucide-react';
import { api, Product } from '../../services/graphqlClient';
import { ProductGrid } from '../catalog/ProductGrid';

interface HomePageProps {
  onAddToCart?: (variantId: string) => void;
}

export const HomePage: React.FC<HomePageProps> = ({ onAddToCart }) => {
  const [featuredProducts, setFeaturedProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const res = await api.searchProducts({ limit: 6 });
        setFeaturedProducts(res.products || []);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div>
      {/* Hero Section */}
      <section style={{ position: 'relative', minHeight: '85vh', display: 'flex', alignItems: 'center', overflow: 'hidden' }}>
        {/* Background Image with Dark Vignette */}
        <div style={{ position: 'absolute', inset: 0, zIndex: 0 }}>
          <img
            src="https://images.unsplash.com/photo-1600210492486-724fe5c67fb0?auto=format&fit=crop&w=2000&q=85"
            alt="StyleHome Atelier"
            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
          />
          <div style={{ position: 'absolute', inset: 0, background: 'linear-gradient(90deg, rgba(13, 15, 18, 0.95) 0%, rgba(13, 15, 18, 0.75) 50%, rgba(13, 15, 18, 0.4) 100%)' }} />
        </div>

        {/* Hero Content */}
        <div className="container" style={{ position: 'relative', zIndex: 1, padding: '4rem 1.5rem', maxWidth: '850px' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.4rem 0.9rem', background: 'rgba(226, 177, 112, 0.15)', border: '1px solid rgba(226, 177, 112, 0.3)', borderRadius: 'var(--radius-full)', color: 'var(--accent-gold)', fontSize: '0.8rem', fontWeight: 600, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: '1.5rem' }}>
            <Sparkles size={14} /> Autumn / Winter 2026 Collection
          </div>

          <h1 className="serif" style={{ fontSize: 'clamp(2.5rem, 6vw, 4.5rem)', fontWeight: 700, lineHeight: 1.1, marginBottom: '1.5rem', color: '#fff' }}>
            Sculptural Simplicity For Intentional Spaces.
          </h1>

          <p style={{ fontSize: '1.15rem', color: 'var(--text-secondary)', lineHeight: 1.7, marginBottom: '2.5rem', maxWidth: '620px' }}>
            Handcrafted Italian velvet sectionals, single-origin oak dining, and precision Scandinavian illumination — engineered for timeless residential architecture.
          </p>

          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            <Link to="/search" className="btn-primary" style={{ padding: '1rem 2rem', fontSize: '1.05rem' }}>
              Explore Collection <ArrowRight size={18} />
            </Link>
            <Link to="/search?category=living-room" className="btn-secondary" style={{ padding: '1rem 2rem', fontSize: '1.05rem' }}>
              <Compass size={18} /> Living Room
            </Link>
          </div>
        </div>
      </section>

      {/* Featured Architecture Showcase */}
      <section className="container" style={{ padding: '6rem 1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '3rem' }}>
          <div>
            <span style={{ fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.12em', color: 'var(--accent-gold)', fontWeight: 600 }}>
              Curated Selection
            </span>
            <h2 className="serif" style={{ fontSize: '2.2rem', fontWeight: 700, marginTop: '0.3rem' }}>
              Featured Atelier Designs
            </h2>
          </div>
          <Link to="/search" style={{ color: 'var(--accent-gold)', fontSize: '0.95rem', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
            View Full Catalog <ArrowRight size={16} />
          </Link>
        </div>

        {loading ? (
          <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--text-secondary)' }}>
            <p>Loading catalog items...</p>
          </div>
        ) : (
          <ProductGrid products={featuredProducts} onAddToCart={onAddToCart} />
        )}
      </section>
    </div>
  );
};
