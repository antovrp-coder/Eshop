import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { SlidersHorizontal, ArrowUpDown } from 'lucide-react';
import { api, SearchResult, Product } from '../../services/graphqlClient';
import { SearchBar } from './SearchBar';
import { ProductGrid } from '../catalog/ProductGrid';

interface SearchPageProps {
  onAddToCart?: (variantId: string) => void;
}

export const SearchPage: React.FC<SearchPageProps> = ({ onAddToCart }) => {
  const [searchParams, setSearchParams] = useSearchParams();
  const queryParam = searchParams.get('q') || '';
  const categoryParam = searchParams.get('category') || '';

  const [searchResult, setSearchResult] = useState<SearchResult | null>(null);
  const [selectedBrands, setSelectedBrands] = useState<string[]>([]);
  const [sortBy, setSortBy] = useState<string>('RELEVANCE');
  const [loading, setLoading] = useState<boolean>(true);

  const fetchResults = async () => {
    setLoading(true);
    try {
      const res = await api.searchProducts({
        query: queryParam,
        categorySlug: categoryParam,
        brands: selectedBrands.length > 0 ? selectedBrands : undefined,
        sortBy,
      });
      setSearchResult(res);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResults();
  }, [queryParam, categoryParam, selectedBrands, sortBy]);

  const handleBrandToggle = (brand: string) => {
    if (selectedBrands.includes(brand)) {
      setSelectedBrands(selectedBrands.filter((b) => b !== brand));
    } else {
      setSelectedBrands([...selectedBrands, brand]);
    }
  };

  const handleSearchQuery = (q: string) => {
    const params = new URLSearchParams(searchParams);
    if (q) params.set('q', q);
    else params.delete('q');
    setSearchParams(params);
  };

  const handleCategorySelect = (catSlug: string) => {
    const params = new URLSearchParams(searchParams);
    if (catSlug) params.set('category', catSlug);
    else params.delete('category');
    setSearchParams(params);
  };

  const brandFacet = searchResult?.facets?.find((f) => f.field === 'brand');
  const categoryFacet = searchResult?.facets?.find((f) => f.field === 'category_slug');

  return (
    <div className="container" style={{ padding: '3rem 1.5rem' }}>
      {/* Top Search Banner */}
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginBottom: '3rem', textAlign: 'center' }}>
        <h1 className="serif" style={{ fontSize: '2.5rem', fontWeight: 700, marginBottom: '0.5rem' }}>
          Explore Curated Catalog
        </h1>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '1.75rem', fontSize: '0.95rem' }}>
          Real-time discovery powered by Elasticsearch full-text index & faceted navigation.
        </p>
        <SearchBar initialValue={queryParam} onSearch={handleSearchQuery} />
      </div>

      {/* Main Content Layout */}
      <div style={{ display: 'grid', gridTemplateColumns: '260px 1fr', gap: '3rem', alignItems: 'start' }}>
        {/* Left Sidebar: Elasticsearch Facet Filters */}
        <aside style={{ background: 'var(--bg-surface)', padding: '1.5rem', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1.5rem', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
            <SlidersHorizontal size={18} color="var(--accent-gold)" />
            <h3 style={{ fontSize: '1rem', fontWeight: 600 }}>Filter & Refine</h3>
          </div>

          {/* Category Filter */}
          <div style={{ marginBottom: '2rem' }}>
            <h4 style={{ fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
              Categories
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <button
                onClick={() => handleCategorySelect('')}
                style={{
                  textAlign: 'left',
                  fontSize: '0.85rem',
                  color: !categoryParam ? 'var(--accent-gold)' : 'var(--text-secondary)',
                  fontWeight: !categoryParam ? 600 : 400,
                }}
              >
                All Categories
              </button>
              {categoryFacet?.buckets.map((b) => (
                <button
                  key={b.key}
                  onClick={() => handleCategorySelect(b.key)}
                  style={{
                    textAlign: 'left',
                    fontSize: '0.85rem',
                    color: categoryParam === b.key ? 'var(--accent-gold)' : 'var(--text-secondary)',
                    fontWeight: categoryParam === b.key ? 600 : 400,
                    display: 'flex',
                    justifyContent: 'space-between',
                  }}
                >
                  <span style={{ textTransform: 'capitalize' }}>{b.key.replace('-', ' ')}</span>
                  <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>({b.count})</span>
                </button>
              ))}
            </div>
          </div>

          {/* Brand Filter */}
          {brandFacet && (
            <div style={{ marginBottom: '1.5rem' }}>
              <h4 style={{ fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
                Atelier Brands
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem' }}>
                {brandFacet.buckets.map((b) => {
                  const isChecked = selectedBrands.includes(b.key);
                  return (
                    <label key={b.key} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '0.85rem', color: 'var(--text-secondary)', cursor: 'pointer' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={() => handleBrandToggle(b.key)}
                          style={{ accentColor: 'var(--accent-gold)' }}
                        />
                        <span>{b.key}</span>
                      </div>
                      <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>({b.count})</span>
                    </label>
                  );
                })}
              </div>
            </div>
          )}
        </aside>

        {/* Right Section: Sort Bar & Product Grid */}
        <section>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '2rem', paddingBottom: '1rem', borderBottom: '1px solid var(--border-subtle)' }}>
            <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
              Showing <strong>{searchResult?.total_count || 0}</strong> architectural designs
            </span>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <ArrowUpDown size={15} color="var(--text-muted)" />
              <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Sort:</span>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                style={{
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-sm)',
                  padding: '0.35rem 0.75rem',
                  fontSize: '0.85rem',
                  outline: 'none',
                }}
              >
                <option value="RELEVANCE">Relevance (BM25)</option>
                <option value="PRICE_ASC">Price: Low to High</option>
                <option value="PRICE_DESC">Price: High to Low</option>
                <option value="RATING">Highest Customer Rating</option>
              </select>
            </div>
          </div>

          {loading ? (
            <div style={{ textAlign: 'center', padding: '5rem 0', color: 'var(--text-secondary)' }}>
              <p>Querying Elasticsearch index...</p>
            </div>
          ) : (
            <ProductGrid products={searchResult?.products || []} onAddToCart={onAddToCart} />
          )}
        </section>
      </div>
    </div>
  );
};
