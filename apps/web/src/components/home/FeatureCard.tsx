import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";

type FeatureCardProps = {
  title: string;
  detail: string;
};

export function FeatureCard({ title, detail }: FeatureCardProps) {
  return (
    <Card className="rounded-2xl border-border/70 bg-card/70">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm">{title}</CardTitle>
        <CardDescription className="text-xs">{detail}</CardDescription>
      </CardHeader>
    </Card>
  );
}
