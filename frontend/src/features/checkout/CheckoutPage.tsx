import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ShieldCheck, CreditCard, CheckCircle2, Lock, ArrowLeft, Loader2 } from 'lucide-react';
import { Cart, api, formatPrice } from '../../services/graphqlClient';

interface CheckoutPageProps {
  cart: Cart | null;
  onOrderSuccess: () => void;
}

export const CheckoutPage: React.FC<CheckoutPageProps> = ({ cart, onOrderSuccess }) => {
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [completedOrder, setCompletedOrder] = useState<any>(null);

  // Address State
  const [formData, setFormData] = useState({
    fullName: 'Jane Doe',
    email: 'jane.doe@stylehome.interior',
    streetLine1: '742 Evergreen Terrace',
    city: 'Springfield',
    stateProvince: 'OR',
    postalCode: '97477',
    countryCode: 'US',
    cardNumber: '4242 •••• •••• 4242',
    expDate: '12/28',
    cvv: '982',
  });

  if (!cart || cart.items.length === 0) {
    if (completedOrder) {
      return (
        <div className="container" style={{ padding: '5rem 1.5rem', maxWidth: '640px', textAlign: 'center' }}>
          <div style={{ padding: '1rem', background: 'rgba(16, 185, 129, 0.1)', borderRadius: 'var(--radius-full)', width: 'fit-content', margin: '0 auto 1.5rem', color: 'var(--accent-emerald)' }}>
            <CheckCircle2 size={48} />
          </div>
          <h1 className="serif" style={{ fontSize: '2.2rem', fontWeight: 700, marginBottom: '0.75rem' }}>
            Order Confirmed!
          </h1>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem', lineHeight: 1.6 }}>
            Thank you for shopping with StyleHome. Your transaction has been securely captured via PCI tokenization.
          </p>

          <div style={{ background: 'var(--bg-surface)', padding: '1.5rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)', textAlign: 'left', marginBottom: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Order ID:</span>
              <span style={{ fontWeight: 600 }}>{completedOrder.id}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Transaction Reference:</span>
              <span style={{ fontWeight: 600 }}>{completedOrder.transaction_id || 'tx_captured'}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Amount Charged:</span>
              <span style={{ fontWeight: 700, color: 'var(--accent-gold)' }}>{formatPrice(completedOrder.grand_total_cents || cart?.grand_total_cents || 0)}</span>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
            <Link to="/orders" className="btn-primary">View in My Orders</Link>
            <Link to="/search" className="btn-secondary">Continue Shopping</Link>
          </div>
        </div>
      );
    }

    return (
      <div className="container" style={{ padding: '6rem 1.5rem', textAlign: 'center' }}>
        <h2 style={{ marginBottom: '1rem' }}>Your Bag is Empty</h2>
        <Link to="/search" className="btn-primary">Explore Products</Link>
      </div>
    );
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setErrorMsg('');

    try {
      const res = await api.placeOrder();
      if (!res.success) {
        setErrorMsg(res.errorMessage || 'Failed to place order.');
      } else {
        setCompletedOrder(res.order);
        onOrderSuccess();
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'An error occurred during order submission.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="container" style={{ padding: '3rem 1.5rem', maxWidth: '1080px' }}>
      <Link to="/search" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '2rem' }}>
        <ArrowLeft size={16} /> Return to Shopping
      </Link>

      <h1 className="serif" style={{ fontSize: '2.4rem', fontWeight: 700, marginBottom: '2rem' }}>
        Secure Express Checkout
      </h1>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))', gap: '3rem', alignItems: 'start' }}>
        {/* Form Column */}
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {/* Shipping Address */}
          <div style={{ background: 'var(--bg-surface)', padding: '1.75rem', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-subtle)' }}>
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '1.25rem' }}>1. Delivery Address</h3>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div style={{ gridColumn: 'span 2' }}>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>Full Name</label>
                <input
                  type="text"
                  required
                  value={formData.fullName}
                  onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
              <div style={{ gridColumn: 'span 2' }}>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>Street Address</label>
                <input
                  type="text"
                  required
                  value={formData.streetLine1}
                  onChange={(e) => setFormData({ ...formData, streetLine1: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>City</label>
                <input
                  type="text"
                  required
                  value={formData.city}
                  onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>Postal Code</label>
                <input
                  type="text"
                  required
                  value={formData.postalCode}
                  onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
            </div>
          </div>

          {/* Payment Method */}
          <div style={{ background: 'var(--bg-surface)', padding: '1.75rem', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem' }}>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>2. Payment Details</h3>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--accent-emerald)', fontSize: '0.8rem' }}>
                <Lock size={14} /> 256-bit Encrypted
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div style={{ gridColumn: 'span 2' }}>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>Card Number</label>
                <div style={{ position: 'relative' }}>
                  <input
                    type="text"
                    required
                    value={formData.cardNumber}
                    onChange={(e) => setFormData({ ...formData, cardNumber: e.target.value })}
                    style={{ width: '100%', padding: '0.65rem 2.5rem 0.65rem 0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                  />
                  <CreditCard size={18} style={{ position: 'absolute', right: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>Expiration</label>
                <input
                  type="text"
                  required
                  value={formData.expDate}
                  onChange={(e) => setFormData({ ...formData, expDate: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>CVV / CVC</label>
                <input
                  type="password"
                  required
                  value={formData.cvv}
                  onChange={(e) => setFormData({ ...formData, cvv: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)' }}
                />
              </div>
            </div>
          </div>

          {errorMsg && (
            <div style={{ padding: '0.85rem 1rem', background: 'rgba(244, 63, 94, 0.1)', border: '1px solid rgba(244, 63, 94, 0.3)', borderRadius: 'var(--radius-md)', color: 'var(--accent-rose)', fontSize: '0.9rem' }}>
              {errorMsg}
            </div>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="btn-primary"
            style={{ width: '100%', padding: '1.1rem', fontSize: '1.05rem' }}
          >
            {submitting ? (
              <>
                <Loader2 size={20} className="animate-spin" /> Orchestrating Saga...
              </>
            ) : (
              <>
                Confirm & Pay {formatPrice(cart.grand_total_cents)}
              </>
            )}
          </button>
        </form>

        {/* Order Review Sidebar */}
        <aside style={{ background: 'var(--bg-surface)', padding: '1.75rem', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-subtle)' }}>
          <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '1.25rem' }}>Order Summary</h3>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
            {cart.items.map((it) => (
              <div key={it.id} style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                <img src={it.image_url} alt="" style={{ width: '50px', height: '50px', objectFit: 'cover', borderRadius: 'var(--radius-sm)' }} />
                <div style={{ flex: 1 }}>
                  <h4 style={{ fontSize: '0.85rem', fontWeight: 600 }}>{it.product_title}</h4>
                  <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Qty: {it.quantity}</p>
                </div>
                <span style={{ fontSize: '0.9rem', fontWeight: 600 }}>{formatPrice(it.unit_price_cents * it.quantity)}</span>
              </div>
            ))}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem', color: 'var(--text-secondary)', borderTop: '1px solid var(--border-subtle)', paddingTop: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>Subtotal</span>
              <span style={{ color: 'var(--text-primary)' }}>{formatPrice(cart.subtotal_cents)}</span>
            </div>
            {cart.discount_cents > 0 && (
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--accent-emerald)' }}>
                <span>Coupon ({cart.coupon?.code})</span>
                <span>-{formatPrice(cart.discount_cents)}</span>
              </div>
            )}
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>Delivery</span>
              <span style={{ color: 'var(--text-primary)' }}>{cart.shipping_cents === 0 ? 'Complimentary' : formatPrice(cart.shipping_cents)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>Tax</span>
              <span style={{ color: 'var(--text-primary)' }}>{formatPrice(cart.tax_cents)}</span>
            </div>
            <div style={{ height: '1px', background: 'var(--border-subtle)', margin: '0.4rem 0' }} />
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-primary)' }}>
              <span>Total</span>
              <span style={{ color: 'var(--accent-gold)' }}>{formatPrice(cart.grand_total_cents)}</span>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
};
