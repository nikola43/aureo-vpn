import React, { createContext, useContext, useState, useEffect } from 'react';
import { api } from '../services/api';
import { User, AuthContextType } from '../types';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [isOperator, setIsOperator] = useState(false);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    const token = localStorage.getItem('access_token');
    if (token) {
      try {
        // Try to fetch dashboard to verify operator status
        await api.getOperatorDashboard();
        setIsOperator(true);
      } catch (error) {
        // Not an operator yet, but still authenticated
        setIsOperator(false);
      }
    }
    setLoading(false);
  };

  const login = async (email: string, password: string) => {
    const response = await api.login({ email, password });
    setUser(response.user);
    await checkAuth();
  };

  const register = async (email: string, password: string, username: string) => {
    const response = await api.register({ email, password, username });
    setUser(response.user);
    await checkAuth();
  };

  const logout = () => {
    api.logout();
    setUser(null);
    setIsOperator(false);
  };

  const value: AuthContextType = {
    user,
    isAuthenticated: !!localStorage.getItem('access_token'),
    isOperator,
    login,
    register,
    logout,
    loading,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
