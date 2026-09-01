import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Navbar } from '../components/layout/Navbar';
import { Footer } from '../components/layout/Footer';
import { CartDrawer } from '../components/cart/CartDrawer';
import { HomePage } from '../features/home/HomePage';
import { SearchPage } from '../features/search/SearchPage';
import { ProductDetailPage } from '../features/catalog/ProductDetailPage';
import { CheckoutPage } from '../features/checkout/CheckoutPage';
import { OrdersPage } from '../features/orders/OrdersPage';
import { api, Cart } from '../services/graphqlClient';

export const App: React.FC = () => {
  const [cart, setCart] = useState<Cart | null>(null);
  const [isCartOpen, setIsCartOpen] = useState<boolean>(false);

  const refreshCart = async () => {
    try {
      const c = await api.getCart();
      setCart(c);
    } catch (err) {
      console.error('Failed to sync cart:', err);
    }
  };

  useEffect(() => {
    refreshCart();
  }, []);

  const handleAddToCart = async (variantId: string, quantity: number = 1) => {
    try {
      const updated = await api.addToCart(variantId, quantity);
      setCart(updated);
      setIsCartOpen(true);
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpdateQuantity = async (cartItemId: string, quantity: number) => {
    try {
      const updated = await api.updateCartItem(cartItemId, quantity);
      setCart(updated);
    } catch (err) {
      console.error(err);
    }
  };

  const handleApplyCoupon = async (code: string) => {
    try {
      const updated = await api.applyCoupon(code);
      setCart(updated);
    } catch (err: any) {
      alert(err.message || 'Invalid coupon');
    }
  };

  return (
    <BrowserRouter>
      <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
        <Navbar cart={cart} onOpenCart={() => setIsCartOpen(true)} />

        <main style={{ flex: 1 }}>
          <Routes>
            <Route path="/" element={<HomePage onAddToCart={handleAddToCart} />} />
            <Route path="/search" element={<SearchPage onAddToCart={handleAddToCart} />} />
            <Route path="/product/:slug" element={<ProductDetailPage onAddToCart={handleAddToCart} />} />
            <Route path="/checkout" element={<CheckoutPage cart={cart} onOrderSuccess={refreshCart} />} />
            <Route path="/orders" element={<OrdersPage />} />
          </Routes>
        </main>

        <Footer />

        <CartDrawer
          isOpen={isCartOpen}
          onClose={() => setIsCartOpen(false)}
          cart={cart}
          onUpdateQuantity={handleUpdateQuantity}
          onApplyCoupon={handleApplyCoupon}
        />
      </div>
    </BrowserRouter>
  );
};
