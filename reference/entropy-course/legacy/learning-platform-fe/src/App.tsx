import { LessonContent } from "./components/LessonContent/LessonContent";
import { Sidebar } from "./components/Sidebar/Sidebar";
import { courseData, currentLessonData } from "./data/courseData";

import "./App.css";

function App() {
  return (
    <div className="app">
      <div className="app__layout">
        <Sidebar course={courseData} currentLessonId={currentLessonData.id} />
        <LessonContent lesson={currentLessonData} />
      </div>
    </div>
  );
}

export default App;
