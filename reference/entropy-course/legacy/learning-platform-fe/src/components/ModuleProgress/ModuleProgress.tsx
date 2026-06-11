import { ModuleProgressData } from "../../types";

import "./ModuleProgress.css";

interface ModuleProgressProps {
  data: ModuleProgressData;
}

export function ModuleProgress({ data }: ModuleProgressProps) {
  return (
    <div className="module-progress">
      <div className="module-progress__header">
        <span className="module-progress__label">Module Progress</span>
        <span className="module-progress__percentage">{data.percentage}%</span>
      </div>
      <div className="module-progress__bar">
        <div className="module-progress__bar-fill" style={{ width: `${data.percentage}%` }} />
      </div>
      <div className="module-progress__count">
        {data.completed}/{data.total} Completed
      </div>
    </div>
  );
}
