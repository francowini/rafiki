import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function HomePage() {
  return (
    <div className="space-y-8">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold">Welcome to Rafiki Thinks</h1>
        <p className="text-xl text-muted-foreground">
          Manage your thoughts, ideas, and reflections
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 max-w-4xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle>Personal</CardTitle>
            <CardDescription>Personal thoughts and reflections</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Work</CardTitle>
            <CardDescription>Work-related notes and ideas</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Ideas</CardTitle>
            <CardDescription>Creative ideas and inspiration</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Learning</CardTitle>
            <CardDescription>Learning notes and insights</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reflection</CardTitle>
            <CardDescription>Deep reflections and analysis</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <div className="text-center">
        <Link href="/thinks">
          <Button size="lg">Get Started</Button>
        </Link>
      </div>
    </div>
  );
}