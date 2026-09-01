import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Star, ShieldCheck, Truck, RotateCcw, ShoppingBag, Check, ArrowLeft } from 'lucide-react';
import { api, Product, formatPrice } from '../../services/graphqlClient';

interface ProductDetailPageProps {
  onAddToCart: (variantId: string, quantity: number) => void;
}

export const ProductDetailPage: React.FC<ProductDetailPageProps> = ({ onAddToCart }) => {
  const { slug } = useParams<{ slug: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [selectedVariantId, setSelectedVariantId] = useState<string>('');
  const [selectedImage, setSelectedImage] = useState<string>('');
  const [quantity, setQuantity] = useState<number>(1);
  const [loading, setLoading] = useState(true);
  const [addedAnimation, setAddedAnimation] = useState(false);

  useEffect(() => {
    async function load() {
      if (!slug) return;
      setLoading(true);
      try {
        const data = await api.getProductBySlug(slug);
        setProduct(data);
        if (data.variants && data.variants.length > 0) {
          setSelectedVariantId(data.variants[0].id);
        }
        setSelectedImage(data.featured_image || (data.images && data.images[0]));
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [slug]);

  if (loading) {
    return (
      <div className="container" style={{ padding: '6rem 0', textAlign: 'center', color: 'var(--text-secondary)' }}>
        <p>Loading curated product details...</p>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="container" style={{ padding: '6rem 0', textAlign: 'center' }}>
        <h2 style={{ marginBottom: '1rem' }}>Product Not Found</h2>
        <Link to="/search" className="btn-primary">Browse Catalog</Link>
      </div>
    );
  }

  const currentVariant = product.variants.find((v) => v.id === selectedVariantId) || product.variants[0];

  const handleAdd = () => {
    if (!currentVariant) return;
    onAddToCart(currentVariant.id, quantity);
    setAddedAnimation(true);
    setTimeout(() => setAddedAnimation(false), 2000);
  };

  return (
    <div className="container" style={{ padding: '3rem 1.5rem' }}>
      <Link to="/search" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '2rem' }}>
        <ArrowLeft size={16} /> Back to Catalog
      </Link>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))', gap: '4rem', alignItems: 'start' }}>
        {/* Left Column: Gallery */}
        <div>
          <div style={{ borderRadius: 'var(--radius-lg)', overflow: 'hidden', height: '480px', border: '1px solid var(--border-subtle)', background: 'var(--bg-surface-elevated)', marginBottom: '1rem' }}>
            <img
              src={selectedImage || currentVariant?.image_url}
              alt={product.title}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
            />
          </div>

          {product.images && product.images.length > 1 && (
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              {product.images.map((img, i) => (
                <button
                  key={i}
                  onClick={() => setSelectedImage(img)}
                  style={{
                    width: '80px',
                    height: '80px',
                    borderRadius: 'var(--radius-md)',
                    overflow: 'hidden',
                    border: selectedImage === img ? '2px solid var(--accent-gold)' : '1px solid var(--border-subtle)',
                  }}
                >
                  <img src={img} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Right Column: Details & Variant Selection */}
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem' }}>
            <span className="badge badge-gold">{product.brand}</span>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', color: 'var(--accent-gold)', fontSize: '0.85rem' }}>
              <Star size={14} fill="currentColor" />
              <span>{product.average_rating} ({product.review_count} Reviews)</span>
            </div>
          </div>

          <h1 className="serif" style={{ fontSize: '2.4rem', fontWeight: 700, lineHeight: 1.2, marginBottom: '1rem' }}>
            {product.title}
          </h1>

          <div style={{ display: 'flex', alignItems: 'baseline', gap: '1rem', marginBottom: '1.5rem' }}>
            <span style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-gold)' }}>
              {formatPrice(currentVariant ? currentVariant.price_cents : product.base_price_cents)}
            </span>
            {currentVariant?.compare_at_price_cents && (
              <span style={{ fontSize: '1.2rem', textDecoration: 'line-through', color: 'var(--text-muted)' }}>
                {formatPrice(currentVariant.compare_at_price_cents)}
              </span>
            )}
          </div>

          <p style={{ color: 'var(--text-secondary)', lineHeight: 1.8, fontSize: '0.95rem', marginBottom: '2rem' }}>
            {product.description}
          </p>

          {/* Variant Selector */}
          {product.variants && product.variants.length > 0 && (
            <div style={{ marginBottom: '2rem' }}>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '0.75rem' }}>
                Select Configuration & Material
              </label>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem' }}>
                {product.variants.map((v) => {
                  const isSelected = v.id === selectedVariantId;
                  return (
                    <button
                      key={v.id}
                      onClick={() => {
                        setSelectedVariantId(v.id);
                        if (v.image_url) setSelectedImage(v.image_url);
                      }}
                      style={{
                        padding: '0.75rem 1.25rem',
                        borderRadius: 'var(--radius-md)',
                        background: isSelected ? 'var(--bg-surface-hover)' : 'var(--bg-surface)',
                        border: isSelected ? '1px solid var(--accent-gold)' : '1px solid var(--border-subtle)',
                        color: isSelected ? 'var(--accent-gold)' : 'var(--text-primary)',
                        fontWeight: 600,
                        fontSize: '0.85rem',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        transition: 'var(--transition-fast)',
                      }}
                    >
                      {isSelected && <Check size={14} />}
                      {v.title}
                    </button>
                  );
                })}
              </div>
            </div>
          )}

          {/* Add to Cart Actions */}
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)', padding: '0.4rem 0.8rem' }}>
              <button onClick={() => setQuantity(Math.max(1, quantity - 1))} style={{ padding: '0.4rem', color: 'var(--text-secondary)' }}>-</button>
              <span style={{ minWidth: '2rem', textAlign: 'center', fontWeight: 600 }}>{quantity}</span>
              <button onClick={() => setQuantity(quantity + 1)} style={{ padding: '0.4rem', color: 'var(--text-secondary)' }}>+</button>
            </div>

            <button
              onClick={handleAdd}
              className="btn-primary"
              style={{ flex: 1, padding: '1rem', fontSize: '1.05rem' }}
            >
              {addedAnimation ? (
                <>
                  <Check size={20} /> Added to Bag
                </>
              ) : (
                <>
                  <ShoppingBag size={20} /> Add to Bag — {formatPrice((currentVariant ? currentVariant.price_cents : product.base_price_cents) * quantity)}
                </>
              )}
            </button>
          </div>

          {/* Guarantees Box */}
          <div style={{ background: 'var(--bg-surface)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)', padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '0.85rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <Truck size={18} color="var(--accent-gold)" />
              <span>Complimentary White-Glove In-Home Delivery</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <RotateCcw size={18} color="var(--accent-gold)" />
              <span>30-Day Risk-Free In-Home Return Trial</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <ShieldCheck size={18} color="var(--accent-gold)" />
              <span>10-Year Structural Hardwood Warranty</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
