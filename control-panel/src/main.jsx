import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ToastContainer } from 'react-toastify'
import 'react-toastify/dist/ReactToastify.css'
import App from './App.jsx'
import './index.css'

// Context
import { AuthProvider } from './context/AuthContext.jsx'

// Pages
import Documentation from './pages/Documentation.jsx'
import ApiReference from './pages/ApiReference.jsx'
import NodeOperators from './pages/NodeOperators.jsx'
import Features from './pages/Features.jsx'
import Downloads from './pages/Downloads.jsx'
import About from './pages/About.jsx'
import Contact from './pages/Contact.jsx'
import PrivacyPolicy from './pages/PrivacyPolicy.jsx'
import TermsOfService from './pages/TermsOfService.jsx'
import Status from './pages/Status.jsx'
import Login from './pages/Login.jsx'
import Register from './pages/Register.jsx'
import Dashboard from './pages/Dashboard.jsx'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <AuthProvider>
        <ToastContainer
          position="top-right"
          autoClose={4000}
          hideProgressBar={false}
          newestOnTop
          closeOnClick
          rtl={false}
          pauseOnFocusLoss
          draggable
          pauseOnHover
          theme="dark"
          toastClassName="!bg-dark-800 !border !border-white/10 !rounded-xl !shadow-xl"
          bodyClassName="!text-gray-200 !font-medium"
          progressClassName="!bg-gold-500"
        />
        <Routes>
          <Route path="/" element={<App />} />
          <Route path="/docs" element={<Documentation />} />
          <Route path="/api" element={<ApiReference />} />
          <Route path="/node-operators" element={<NodeOperators />} />
          <Route path="/features" element={<Features />} />
          <Route path="/downloads" element={<Downloads />} />
          <Route path="/about" element={<About />} />
          <Route path="/contact" element={<Contact />} />
          <Route path="/privacy" element={<PrivacyPolicy />} />
          <Route path="/terms" element={<TermsOfService />} />
          <Route path="/status" element={<Status />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/dashboard" element={<Dashboard />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
