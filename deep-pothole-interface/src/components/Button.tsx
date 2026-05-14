import { Link } from "react-router";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
  to?: string;
}

function Button({
  label,
  to,
  type = "button",
  className,
  ...props
}: ButtonProps) {
  const baseStyles =
    "flex items-center justify-center w-full bg-yellow-primary hover:bg-[#e0b441] text-blue-secondary font-bold text-xl py-4 rounded-[25px] transition-colors shadow-lg cursor-pointer text-center";

  if (to) {
    return (
      <Link to={to} className={`${baseStyles} ${className || ""}`}>
        {label}
      </Link>
    );
  }

  return (
    <button
      type={type}
      className={`${baseStyles} ${className || ""}`}
      {...props}
    >
      {label}
    </button>
  );
}

export default Button;
