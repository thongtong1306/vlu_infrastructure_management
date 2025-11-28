// index.js
import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Provider, useSelector, useDispatch } from 'react-redux';
import { createStore } from 'redux';

import './pages/main.scss';

import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import Dashboard from './pages/Dashboard';
import BorrowItem from './pages/BorrowItem';
import AddItem from './pages/AddItem';
import RegisterPage from "./pages/RegisterPage";
import LabsHome from './pages/LabsHome';
import LabMain from './pages/labs/LabMain';
import LabTwo from './pages/labs/LabTwo';
import LabThree from './pages/labs/LabThree';
import Equipments from "./pages/Equipments";
import InstructionView from "./pages/InstructionView";

// ----- minimal Redux -----
const initialState = {
    session: null,        // { token, exp, user }
    skuInfos: {},
    subscriptionInfo: null,
    num: 0,
};

function reducer(state = initialState, action) {
    switch (action.type) {
        case 'GENERAL_VALUE_UPDATE':
            return { ...state, [action.key]: action.value };
        default:
            return state;
    }
}
const store = createStore(
    reducer,
    // DevTools (optional)
    typeof window !== 'undefined' && window.__REDUX_DEVTOOLS_EXTENSION__
        ? window.__REDUX_DEVTOOLS_EXTENSION__()
        : undefined
);

// helper so class pages can call onReduxUpdate (matches your redux-mapping.js)
export function useReduxMapping() {
    const dispatch = useDispatch();
    return {
        onReduxUpdate: (key, val) =>
            new Promise((resolve) => {
                dispatch({ type: 'GENERAL_VALUE_UPDATE', key, value: val });
                resolve();
            }),
    };
}

// protect routes that need a session
function Protected({ children }) {
    const session = useSelector((s) => s.session);
    return session?.token ? children : <Navigate to="/login" replace />;
}

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="/labs" element={<LabsHome />} />
                <Route path="/labs/main" element={<LabMain />} />
                <Route path="/labs/lab-2" element={<LabTwo />} />
                <Route path="/labs/lab-3" element={<LabThree />} />
                <Route path="/equipments" element={<Equipments />} />
                <Route path="/instructions/:id" element={<InstructionView />} />
                <Route path="/dashboard"
                    element={
                        <Dashboard />
                    }
                />
                <Route path="/borrow"
                    element={
                        <Protected>
                            <BorrowItem />
                        </Protected>
                    }
                />
                <Route path="/add-item"
                    element={
                        <Protected>
                            <AddItem />
                        </Protected>
                    }
                />
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </BrowserRouter>
    );
}

const root = createRoot(document.getElementById('root'));
root.render(
    <Provider store={store}>
        <App />
    </Provider>
);
