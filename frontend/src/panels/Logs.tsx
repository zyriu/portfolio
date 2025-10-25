import React from "react";
import LogPanel from "../components/LogPanel";

export default function Logs({ initialJobName }: { initialJobName?: string }) {
  return <LogPanel initialJobName={initialJobName} />;
}
