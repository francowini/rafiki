import { Home, Heart, Sparkles, Brain, Clock } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export interface NavItem {
  label: string;
  icon: LucideIcon;
  href: string;
  color: string;
}

export const navItems: NavItem[] = [
  { label: 'Dashboard', icon: Home, href: '/', color: 'text-gray-700' },
  { label: 'Values', icon: Heart, href: '/values', color: 'text-rose-600' },
  { label: 'Life Visions', icon: Sparkles, href: '/life-visions', color: 'text-purple-600' },
  { label: 'Thinks', icon: Brain, href: '/thinks', color: 'text-blue-600' },
  { label: 'Moments', icon: Clock, href: '/momentos', color: 'text-teal-600' },
];
