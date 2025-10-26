import React, { useState, useEffect } from "react";
import GlobalSettings from "../panels/GlobalSettings";

export type Section = "jobs" | "logs" | "settings";

function getSectionTitle(section: Section): string {
  switch (section) {
    case "jobs": return "Jobs";
    case "logs": return "Logs";
    case "settings": return "Settings";
    default: return "Portfolio Manager";
  }
}

export function Layout({
  section,
  setSection,
  children,
}: {
  section: Section;
  setSection: (s: Section) => void;
  children: React.ReactNode;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsClosing, setSettingsClosing] = useState(false);

  const closeMenu = () => {
    setMenuOpen(false);
  };

  const navigateTo = (s: Section) => {
    setSection(s);
    closeMenu();
  };

  const toggleSettings = () => {
    if (settingsOpen) {
      // Start closing animation
      setSettingsClosing(true);
    } else {
      // Open settings
      setSettingsOpen(true);
      setSettingsClosing(false);
    }
  };

  // Handle closing animation completion
  useEffect(() => {
    if (settingsClosing) {
      const timer = setTimeout(() => {
        setSettingsOpen(false);
        setSettingsClosing(false);
      }, 300); // Match animation duration
      return () => clearTimeout(timer);
    }
  }, [settingsClosing]);

  return (
    <div className="app-container">
      {/* Header */}
      <header className="app-header">
        <button 
          className="burger-btn" 
          onClick={() => setMenuOpen(!menuOpen)}
        >
          ☰
        </button>
        
        <div className="app-title">{getSectionTitle(section)}</div>
        
        <button 
          className="burger-btn" 
          onClick={toggleSettings}
        >
          ⚙
        </button>
      </header>

      {/* Navigation Menu (Left) */}
      {menuOpen && (
        <>
          <div className="overlay" onClick={closeMenu} />
          <nav className="menu-drawer menu-drawer-left">
            <div className="menu-header">Navigation</div>
            <MenuItem 
              label="Jobs" 
              active={section === "jobs"} 
              onClick={() => navigateTo("jobs")} 
            />
            <MenuItem 
              label="Logs" 
              active={section === "logs"} 
              onClick={() => navigateTo("logs")} 
            />
          </nav>
        </>
      )}

      {/* Settings Panel (Right) */}
      {settingsOpen && (
        <>
          <div className="overlay" onClick={toggleSettings} />
          <div className={`settings-panel ${settingsClosing ? 'closing' : ''}`}>
            <div className="settings-header">
              <h2>Settings</h2>
              <button className="settings-close-btn" onClick={toggleSettings}>
                ✕
              </button>
            </div>
            <div className="settings-content">
              <GlobalSettings />
            </div>
          </div>
        </>
      )}

      {/* Main Content */}
      <main className="app-content">
        {children}
      </main>
    </div>
  );
}

function MenuItem({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`menu-item ${active ? "menu-item-active" : ""}`}
    >
      {label}
    </button>
  );
}
