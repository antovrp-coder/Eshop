import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { PackageCheck, Clock, CheckCircle2, Truck, ExternalLink } from 'lucide-react';
import { api, Order, formatPrice } from '../../services/graphqlClient';

export const OrdersPage: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    async function load() {
      try {
        const list = await api.getMyOrders();
        setOrders(list);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="container" style={{ padding: '3rem 1.5rem', maxWidth: '960px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
        <PackageCheck size={28} color="var(--accent-gold)" />
        <h1 className="serif" style={{ fontSize: '2.2rem', fontWeight: 700 }}>Your Orders</h1>
      </div>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '2.5rem', fontSize: '0.95rem' }}>
        Real-time distributed fulfillment tracking across PostgreSQL and Kafka event streams.
      </p>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--text-secondary)' }}>
          <p>Retrieving order ledger...</p>
        </div>
      ) : orders.length === 0 ? (
        <div style={{ background: 'var(--bg-surface)', borderRadius: 'var(--radius-lg)', padding: '4rem 2rem', textAlign: 'center', border: '1px solid var(--border-subtle)' }}>
          <Clock size={40} style={{ margin: '0 auto 1rem', opacity: 0.4 }} />
          <h3 style={{ fontSize: '1.2rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Orders Found</h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>You haven't placed any orders yet.</p>
          <Link to="/search" className="btn-primary">Browse Design Catalog</Link>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {orders.map((ord) => (
            <div
              key={ord.id}
              style={{
                background: 'var(--bg-surface)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border-subtle)',
                overflow: 'hidden',
              }}
            >
              {/* Card Header */}
              <div style={{ background: 'var(--bg-surface-elevated)', padding: '1rem 1.5rem', borderBottom: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                <div style={{ display: 'flex', gap: '2rem' }}>
                  <div>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Order Placed</span>
                    <p style={{ fontSize: '0.85rem', fontWeight: 600 }}>{new Date(ord.created_at || Date.now()).toLocaleDateString()}</p>
                  </div>
                  <div>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Order Number</span>
                    <p style={{ fontSize: '0.85rem', fontWeight: 600 }}>{ord.order_number || ord.id}</p>
                  </div>
                  <div>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Total Amount</span>
                    <p style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--accent-gold)' }}>{formatPrice(ord.grand_total_cents)}</p>
                  </div>
                </div>

                <div>
                  <span className="badge badge-emerald" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}>
                    <CheckCircle2 size={13} /> {ord.status || 'CONFIRMED'}
                  </span>
                </div>
              </div>

              {/* Items in Order */}
              <div style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {ord.items && ord.items.map((it, idx) => (
                  <div key={idx} style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                    <img src={it.image_url} alt="" style={{ width: '64px', height: '64px', objectFit: 'cover', borderRadius: 'var(--radius-sm)' }} />
                    <div style={{ flex: 1 }}>
                      <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>{it.product_title}</h4>
                      <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Quantity: {it.quantity} • {formatPrice(it.unit_price_cents)} each</p>
                    </div>
                    <span style={{ fontSize: '0.95rem', fontWeight: 600 }}>{formatPrice(it.total_price_cents || (it.unit_price_cents * it.quantity))}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
