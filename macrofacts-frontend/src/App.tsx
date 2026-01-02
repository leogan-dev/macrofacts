import { BrowserRouter, Navigate, Route, Routes, useNavigate } from "react-router-dom";
import { AuthPage } from "./pages/AuthPage";
import TodayPage from "./pages/TodayPage";
import { clearToken, getToken } from "./api/client";
import AddFoodPage from "./pages/AddFoodPage";


function RequireAuth({ children }: { children: JSX.Element }) {
    const token = getToken();
    if (!token) return <Navigate to="/auth" replace />;
    return children;
}

function AuthedToday() {
    const nav = useNavigate();
    return (
        <TodayPage
            onLogout={() => {
                clearToken();
                nav("/auth", { replace: true });
            }}
        />
    );
}

function AuthRoute() {
    const nav = useNavigate();
    return <AuthPage onAuthed={() => nav("/today", { replace: true })} />;
}

export default function App() {
    const authed = !!getToken();
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/auth" element={<AuthRoute />} />
                <Route
                    path="/today"
                    element={
                        <RequireAuth>
                            <AuthedToday />
                        </RequireAuth>
                    }
                />
                <Route path="/" element={<Navigate to={authed ? "/today" : "/auth"} replace />} />
                <Route path="*" element={<Navigate to="/" replace />} />
                <Route path="/add-food" element={<AddFoodPage />} />
            </Routes>
        </BrowserRouter>
    );
}
