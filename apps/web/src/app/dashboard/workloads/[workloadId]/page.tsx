import { WorkloadDetail } from "@/components/dashboard/workload-detail";
export default async function WorkloadPage({ params }: { params: Promise<{ workloadId: string }> }) { return <WorkloadDetail id={(await params).workloadId} />; }
