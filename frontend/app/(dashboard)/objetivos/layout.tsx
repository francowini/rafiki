import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Objetivos | Rafiki',
  description: 'Gestiona tus objetivos y mide tu progreso hacia tus visiones de vida',
};

export default function ObjetivosLayout({ children }: { children: React.ReactNode }) {
  return children;
}
