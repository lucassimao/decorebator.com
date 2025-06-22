import React from 'react';

interface TranslatedListProps {
  items: string | string[];
}

export const TranslatedList: React.FC<TranslatedListProps> = ({ items }) => {
  // Handle both string and array types
  const itemArray = Array.isArray(items) ? items : [items];
  
  return (
    <>
      {itemArray.map((item, index) => (
        <li key={index}>{item}</li>
      ))}
    </>
  );
};