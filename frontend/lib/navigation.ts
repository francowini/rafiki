import { Home, Heart, Sparkles, Brain, Clock, Users } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export interface NavItem {
  label: string;
  icon: LucideIcon;
  href: string;
  color: string;
}

export const navItems: NavItem[] = [
  { label: 'Panel', icon: Home, href: '/', color: 'text-gray-700' },
  { label: 'Valores', icon: Heart, href: '/values', color: 'text-rose-600' },
  { label: 'Roles', icon: Users, href: '/roles', color: 'text-purple-600' },
  { label: 'Visiones de Vida', icon: Sparkles, href: '/life-visions', color: 'text-violet-600' },
  { label: 'Pensamientos', icon: Brain, href: '/thinks', color: 'text-blue-600' },
  { label: 'Momentos', icon: Clock, href: '/momentos', color: 'text-teal-600' },
];
