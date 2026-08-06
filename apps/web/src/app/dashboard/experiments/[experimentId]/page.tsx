import { ExperimentDetail } from "@/components/dashboard/experiment-detail";
export default async function ExperimentPage({params}:{params:Promise<{experimentId:string}>}){return <ExperimentDetail id={(await params).experimentId}/>;}
