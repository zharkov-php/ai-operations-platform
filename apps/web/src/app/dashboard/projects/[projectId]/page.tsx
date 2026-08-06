import { ProjectDetail } from "@/components/dashboard/project-detail";
export default async function ProjectPage({ params }: { params: Promise<{ projectId: string }> }) { return <ProjectDetail id={(await params).projectId} />; }
