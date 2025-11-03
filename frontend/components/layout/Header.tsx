import Link from "next/link";
import { Button } from "@/components/ui/button";

export function Header() {
  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold">
            Rafiki Thinks
          </Link>
          <nav className="flex gap-4">
            <Link href="/thinks">
              <Button variant="ghost">Thinks</Button>
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">
            {process.env.NEXT_PUBLIC_ENV || "development"}
          </span>
        </div>
      </div>
    </header>
  );
}