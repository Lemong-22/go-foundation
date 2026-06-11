import "./Header.css";

export function Header() {
  return (
    <header className="header">
      <div className="header__top">
        <div className="header__avatar">👤</div>
        <button className="header__icon-btn" aria-label="Notifications">
          🔔
        </button>
      </div>
      <div className="header__search">
        <span className="header__search-icon">🔍</span>
        <input
          type="text"
          className="header__search-input"
          placeholder="Search courses, modules..."
        />
      </div>
    </header>
  );
}
