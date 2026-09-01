import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Sparkles } from 'lucide-react';
import { api } from '../../services/graphqlClient';

interface SearchBarProps {
  initialValue?: string;
  onSearch?: (query: string) => void;
}

export const SearchBar: React.FC<SearchBarProps> = ({ initialValue = '', onSearch }) => {
  const [query, setQuery] = useState(initialValue);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const navigate = useNavigate();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const timer = setTimeout(async () => {
      if (query.trim().length >= 1) {
        try {
          const list = await api.getSuggestions(query);
          setSuggestions(list);
        } catch (e) {
          console.error(e);
        }
      } else {
        setSuggestions([]);
      }
    }, 150);

    return () => clearTimeout(timer);
  }, [query]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setShowSuggestions(false);
    if (onSearch) {
      onSearch(query);
    } else {
      navigate(`/search?q=${encodeURIComponent(query)}`);
    }
  };

  const handleSelectSuggestion = (text: string) => {
    setQuery(text);
    setShowSuggestions(false);
    if (onSearch) {
      onSearch(text);
    } else {
      navigate(`/search?q=${encodeURIComponent(text)}`);
    }
  };

  return (
    <div ref={containerRef} style={{ position: 'relative', width: '100%', maxWidth: '640px' }}>
      <form onSubmit={handleSubmit} style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
        <input
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          placeholder="Search Scandinavian lighting, velvet sofas, oak tables..."
          style={{
            width: '100%',
            padding: '1rem 3.5rem 1rem 1.25rem',
            background: 'var(--bg-surface-elevated)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-lg)',
            fontSize: '0.95rem',
            boxShadow: 'var(--shadow-subtle)',
            outline: 'none',
            transition: 'var(--transition-fast)',
          }}
        />
        <button
          type="submit"
          style={{
            position: 'absolute',
            right: '0.75rem',
            padding: '0.65rem 1rem',
            background: 'var(--accent-gold)',
            color: '#000',
            borderRadius: 'var(--radius-md)',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
          }}
        >
          <Search size={17} />
        </button>
      </form>

      {/* Autocomplete Suggestions Popup */}
      {showSuggestions && suggestions.length > 0 && (
        <div
          className="glass-panel"
          style={{
            position: 'absolute',
            top: 'calc(100% + 8px)',
            left: 0,
            right: 0,
            borderRadius: 'var(--radius-md)',
            boxShadow: 'var(--shadow-elevated)',
            zIndex: 60,
            overflow: 'hidden',
            padding: '0.5rem 0',
          }}
        >
          <div style={{ padding: '0.4rem 1rem', fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-muted)', letterSpacing: '0.08em', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <Sparkles size={13} color="var(--accent-gold)" /> Instant Suggestions (Elasticsearch)
          </div>
          {suggestions.map((item, idx) => (
            <button
              key={idx}
              onClick={() => handleSelectSuggestion(item)}
              style={{
                width: '100%',
                padding: '0.65rem 1rem',
                textAlign: 'left',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                fontSize: '0.9rem',
                color: 'var(--text-primary)',
                transition: 'var(--transition-fast)',
              }}
              onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-surface-hover)')}
              onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
            >
              <Search size={15} color="var(--text-muted)" />
              <span>{item}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
