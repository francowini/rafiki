'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';

// Types
export interface User {
  sub: string;        // User ID from JWT
  email: string;      // Email from JWT
  name: string;       // Display name from JWT
  roles: string[];    // User roles from JWT
}

export interface DecodedJWT {
  sub: string;
  email: string;
  name: string;
  roles: string[];
  exp: number;
  iat: number;
  iss: string;
}

export interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

// Context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider
interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Initialize auth state from localStorage
  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token');

    if (storedToken) {
      try {
        const decoded = decodeJWT(storedToken);

        // Check if token is expired
        if (decoded.exp * 1000 < Date.now()) {
          // Token expired, clear it
          localStorage.removeItem('auth_token');
          setIsLoading(false);
          return;
        }

        // Token valid, set auth state
        setToken(storedToken);
        setUser({
          sub: decoded.sub,
          email: decoded.email,
          name: decoded.name,
          roles: decoded.roles,
        });
      } catch (error) {
        console.error('Failed to decode token:', error);
        localStorage.removeItem('auth_token');
      }
    }

    setIsLoading(false);
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';
      const kid = process.env.NEXT_PUBLIC_AUTH_KID || '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1';

      // Create Basic Auth header
      const credentials = btoa(`${email}:${password}`);

      const response = await fetch(`${apiUrl}/v1/auth/token/${kid}`, {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${credentials}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Login failed');
      }

      const data = await response.json();
      const jwtToken = data.token;

      // Decode JWT to get user info
      const decoded = decodeJWT(jwtToken);

      // Store token
      localStorage.setItem('auth_token', jwtToken);

      // Update state
      setToken(jwtToken);
      setUser({
        sub: decoded.sub,
        email: decoded.email,
        name: decoded.name,
        roles: decoded.roles,
      });

      // Redirect to dashboard
      router.push('/');
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  // Logout function
  const logout = () => {
    localStorage.removeItem('auth_token');
    setToken(null);
    setUser(null);
    router.push('/login');
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!token && !!user,
    isLoading,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Helper to decode JWT (client-side)
function decodeJWT(token: string): DecodedJWT {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );

    return JSON.parse(jsonPayload);
  } catch (error) {
    throw new Error('Invalid token format');
  }
}
