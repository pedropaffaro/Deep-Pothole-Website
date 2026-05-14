import { Routes, Route } from "react-router";
import Home from "./pages/Home";
import Complaint from "./pages/Complaint";
import NotFound from "./pages/NotFound";

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/complaint" element={<Complaint />} />

      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export default App;
